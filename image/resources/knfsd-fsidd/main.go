package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/pflag"
)

func main() {
	var err error

	cfg := new(Config)
	f := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	// setup flags before reading the config files , otherwise the pflag package
	// will overwrite the config with the default values
	f.StringVar(&cfg.SocketPath, "socket", defaultSocketPath, "")
	f.StringVar(&cfg.Database.URL, "database-url", "", "")
	f.StringVar(&cfg.Database.Instance, "database-instance", "", "")
	f.StringVar(&cfg.Database.TableName, "table-name", "", "")
	f.BoolVar(&cfg.Database.IAMAuth, "iam-auth", false, "")
	f.BoolVar(&cfg.Database.PrivateIP, "private-ip", false, "")
	f.BoolVar(&cfg.Debug, "debug", false, "")

	// read the config file before parsing the command line arguments so
	// that the command line arguments override any config values
	err = readDefaultConfig(cfg)
	if err != nil {
		log.Error.Printf("could not read config: %s", err)
		os.Exit(2)
	}

	// override values from the config file with environment variables
	err = readEnv(cfg)
	if err != nil {
		printConfigError(err)
		os.Exit(2)
	}

	// command line arguments overrides all other sources
	err = f.Parse(os.Args[1:])
	if errors.Is(err, pflag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		log.Error.Print(err)
		os.Exit(2)
	}

	if cfg.Debug {
		log.EnableDebug()
	}

	err = cfg.Validate()
	if err != nil {
		printConfigError(err)
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = run(ctx, cfg)
	if err != nil {
		log.Error.Print(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *Config) error {
	var err error

	db, err := connect(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	f := FSIDSource{
		db:        db,
		tableName: cfg.Database.TableName,
	}

	if cfg.Database.CreateTable {
		err = f.CreateTable(ctx)
		if err != nil {
			return err
		}
	}

	s, err := resolveSocket(cfg.SocketPath)
	if err != nil {
		return err
	}
	defer s.Close()

	s.Handle("get_fsidnum", func(ctx context.Context, path string) (string, error) {
		if path == "" {
			return "", ErrInvalidArgument
		}

		var fsid int32
		found := true

		err := withRetry(ctx, func() error {
			var err error
			fsid, err = f.GetFSID(ctx, path)
			if errors.Is(err, pgx.ErrNoRows) {
				err = nil
				found = false
			}
			return err
		})

		if err != nil {
			return "", err
		} else if found {
			return strconv.FormatInt(int64(fsid), 10), nil
		} else {
			return "", nil
		}
	})

	s.Handle("get_or_create_fsidnum", func(ctx context.Context, path string) (string, error) {
		if path == "" {
			return "", ErrInvalidArgument
		}

		var fsid int32
		err := withRetry(ctx, func() error {
			var err error
			fsid, err = f.GetFSID(ctx, path)
			if errors.Is(err, pgx.ErrNoRows) {
				// FSID not found for path, so try and allocate one.
				// This might fail with a 23505 unique_violation if the path has
				// already been allocated an FSID by different process. withRetry
				// will then retry this whole block and will find the FSID
				// allocated by the other process.
				fsid, err = f.AllocateFSID(ctx, path)
			}
			return err
		})
		return strconv.FormatInt(int64(fsid), 10), err
	})

	s.Handle("get_path", func(ctx context.Context, arg string) (string, error) {
		fsid, err := strconv.ParseInt(arg, 10, 32)
		if err != nil {
			return "", ErrInvalidArgument
		}
		if fsid < 1 {
			return "", ErrInvalidArgument
		}

		var path string
		err = withRetry(ctx, func() error {
			var err error
			path, err = f.GetPath(ctx, int32(fsid))
			if errors.Is(err, pgx.ErrNoRows) {
				err = nil
				path = ""
			}
			return err
		})

		return path, err
	})

	s.Handle("version", func(ctx context.Context, arg string) (string, error) {
		return "1", nil
	})

	go func() {
		<-ctx.Done()
		_, err := daemon.SdNotify(false, daemon.SdNotifyStopping)
		if err != nil {
			log.Error.Print(err)
		}

		deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		s.Shutdown(deadline)
		s.Close()
	}()

	_, err = daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		return err
	}
	log.Info.Print("fsidd service started")

	err = s.Serve()
	if errors.Is(err, ErrServerClosed) {
		err = nil
	}
	return err
}
