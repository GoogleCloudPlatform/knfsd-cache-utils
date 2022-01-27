package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/collectd"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/mounts"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/slab"
	"github.com/prometheus/procfs"
)

var mountPoint = flag.String("proc", procfs.DefaultMountPoint, "path to procfs filesystem")
var socket = flag.String("socket", "", "path to collectd unix socket")
var simulate = flag.String("simulate", "", "path to directory containing snapshots from procfs")
var mode = flag.String("mode", "proxy", "Sets which metrics are collected, valid options are proxy or client")

func run(ctx context.Context, report func() error) {
	ticker := time.NewTicker(collectd.Interval())

	for {
		err := report()
		if err != nil {
			log.Printf("ERROR: %s", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func runSimulation() {
	log.SetFlags(0)
	base := *simulate

	dirs, err := os.ReadDir(base)
	if err != nil {
		log.Fatal(err)
	}

	names := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		names = append(names, d.Name())
	}

	// try to sort the directories numerically, fallback to lexicographic
	sort.Slice(names, func(i, j int) bool {
		a := names[i]
		b := names[j]
		x, e1 := strconv.Atoi(a)
		y, e2 := strconv.Atoi(b)
		if e1 == nil && e2 == nil {
			return x < y
		}
		return a < b
	})

	divider := strings.Repeat("=", 72)
	_ = divider

	for _, name := range names {
		// fmt.Printf("> %s\n", name)
		simulateStep(path.Join(base, name))
		// fmt.Println(divider)
	}
	fmt.Printf("Done")
}

func simulateStep(path string) {
	ctx := context.Background()
	fs, err := procfs.NewFS(path)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}

	p, err := fs.Self()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}

	err = slab.Report(ctx, fs)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}

	err = mounts.Report(ctx, p)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
}

func main() {
	flag.BoolVar(&collectd.EnableMeta, "enable-meta", true, "Enable setting metadata when sending metrics. Requires collectd 5.11.0 or greater.")

	stringFromEnv(socket, "METRICS_SOCKET_PATH")
	stringFromEnv(mode, "METRICS_MODE")
	boolFromEnv(&collectd.EnableMeta, "METRICS_ENABLE_META")

	flag.Parse()

	switch *mode {
	case "proxy":
	case "client":
	default:
		log.Fatalf("invalid mode \"%s\", valid options are \"proxy\" or \"client\".\n", *mode)
	}

	if *socket != "" {
		collectd.UseUnixSocket(*socket)
	}

	if *simulate != "" {
		runSimulation()
		return
	}

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	fs, err := procfs.NewFS(*mountPoint)
	if err != nil {
		log.Fatal(err)
	}

	p, err := fs.Self()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: investigate procfs/nfs client and server metrics

	if *mode == "proxy" {
		go run(ctx, func() error {
			count, err := countConnectedClients()
			if err != nil {
				return annotate(err, "could not report client count")
			}
			connectedClients.Write(ctx, time.Now(), float64(count))
			return nil
		})
	}

	go run(ctx, func() error {
		err := slab.Report(ctx, fs)
		return annotate(err, "could not report cache metrics")
	})

	// run the last metric on the main goroutine instead of a background goroutine
	run(ctx, func() error {
		err := mounts.Report(ctx, p)
		return annotate(err, "could not report NFS mount metrics")
	})
}

// stringFromEnv sets a string flag from an environment variable if it was not
// set by the command line. Must call before flag.Parse so that explicit command
// line arguments take precedence over environment variables.
// Must call after defining the flags though, so that environment variables can
// override defaults.
func stringFromEnv(p *string, name string) {
	val, exists := os.LookupEnv(name)
	if exists && val != "" {
		*p = val
	}
}

func boolFromEnv(p *bool, name string) {
	s, exists := os.LookupEnv(name)
	if !exists || s == "" {
		return
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Printf("invalid value for %s: could not parse '%s' as a boolean: %s", name, s, err)
		os.Exit(2) // same exit code as args.Parse
	}

	*p = b
}

func annotate(err error, msg string) error {
	if err == nil {
		return nil
	} else {
		return fmt.Errorf("%s: %w", msg, err)
	}
}
