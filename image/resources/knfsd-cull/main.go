/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-watchdog/internal/log"
)

const (
	dropCachesPath      = "/proc/sys/vm/drop_caches"
	confPath            = "/etc/knfsd-cull.conf"
	cachefilesdConfPath = "/etc/cachefilesd.conf"
)

type config struct {
	cacheRoot   string
	threshold   uint64
	lastAccess  time.Duration
	interval    time.Duration
	quietPeriod time.Duration
}

func main() {
	var debug, now bool
	var cfg config
	var err error

	flag.BoolVar(&debug, "debug", false, "Enable debug level logging.")
	flag.BoolVar(&now, "now", false, "Run a single cull immediately, then terminate.")
	flag.Parse()

	cfg, err = readConfig(confPath)
	if err != nil {
		log.Fatalf("Could not parse %s: %s", confPath, err)
	}

	// Read the cache root from the cachefilesd.conf file to avoid duplicating
	// this information.
	cfg.cacheRoot, err = readCacheRoot(cachefilesdConfPath)
	if err != nil {
		log.Fatalf("Could not parse %s: %s", cachefilesdConfPath, err)
	}

	if debug {
		log.EnableDebug()
	}

	err = validateCanWrite(dropCachesPath)
	if err != nil {
		log.Fatalf("Could not open %s for writing\n", dropCachesPath)
	}

	err = validateCacheRoot(cfg.cacheRoot)
	if err != nil {
		log.Fatalf("invalid cache-path '%s': %s", cfg.cacheRoot, err)
	}

	_, err = statfs(cfg.cacheRoot)
	if err != nil {
		log.Fatalf("could not statfs '%s': '%s'", cfg.cacheRoot, err)
	}

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	cull := culler{
		config: cfg,
	}
	if now {
		cull.run(ctx)
	} else {
		cull.watch(ctx)
	}
}

func validateCanWrite(path string) error {
	f, err := os.OpenFile(dropCachesPath, os.O_WRONLY|os.O_APPEND, 0)
	if err == nil {
		f.Close()
	}
	return err
}

func validateCacheRoot(cacheRoot string) error {
	if cacheRoot == "" {
		return fmt.Errorf("required")
	}

	if err := validateDirectoryExists(path.Join(cacheRoot, "cache")); err != nil {
		return err
	}

	if err := validateDirectoryExists(path.Join(cacheRoot, "graveyard")); err != nil {
		return err
	}

	return nil
}

func validateDirectoryExists(name string) error {
	s, err := os.Stat(name)
	if err != nil {
		return err
	}

	if !s.IsDir() {
		return fmt.Errorf("could not find directory '%s'", name)
	}

	return nil
}
