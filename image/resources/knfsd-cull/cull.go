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
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-watchdog/internal/log"
	"golang.org/x/sys/unix"
)

type culler struct {
	config
}

// maximum time value, any other time will have to be less than this value
var maxTime = (time.Time{}).Add(time.Duration(1<<63 - 1))

func (cfg *culler) watch(ctx context.Context) {
	running := true
	for running {
		nextCull := cfg.run(ctx)
		// Using sleep instead of a ticker because the interval represents the
		// minimum interval to wait between culling checks.
		running = sleep(ctx, nextCull)
	}
}

func (cfg *culler) run(ctx context.Context) time.Duration {
	shouldCull, err := cfg.checkFreeSpace()
	if err != nil {
		log.Error(err)
		return cfg.interval
	}

	if !shouldCull {
		// culling not required, so wait for the check interval
		return cfg.interval
	}

	culledSomething, oldestFile := cfg.cullFiles(ctx)
	if culledSomething {
		dropVMCache()
	}

	// Even if nothing was culled, we still had to walk the filesystem to
	// discover that, so wait for the next cull check based on the quite
	// period.
	return cfg.nextCullCheck(oldestFile)
}

func (cfg *culler) checkFreeSpace() (bool, error) {
	s, err := statfs(cfg.cacheRoot)
	if err != nil {
		return false, err
	}

	// Divide total by 100 to calculate as a percentage without using floating
	// point. This would fail if blocks or files is less than 100, but the
	// cache isn't going to be very useful on a disk with only 100 bytes, or
	// 100 inodes.
	blocks := s.Bavail / (s.Blocks / 100)
	files := s.Ffree / (s.Files / 100)
	shouldCull := blocks < cfg.threshold || files < cfg.threshold

	log.Debugf("files : %d of %d (%d%%)", s.Ffree, s.Files, files)
	log.Debugf("blocks: %d of %d (%d%%)", s.Bavail, s.Blocks, blocks)

	return shouldCull, nil
}

func (cfg *culler) cullFiles(ctx context.Context) (bool, time.Time) {
	dir := path.Join(cfg.cacheRoot, "cache")
	now := time.Now()
	cutoff := now.Add(-cfg.lastAccess)

	culledSomething := false
	filesCulled := 0
	totalFiles := 0

	// Track the oldest (non-culled) file in the cache.
	// Initialize this to the largest time value so that any file will be less
	// than this value.
	oldestFile := maxTime

	log.Infof("cull: removing files older than %s", cutoff)

	// The callback handles (logs) all the errors internally, so we don't
	// care about the error returned by WalkDir.
	// Even if the context was cancelled, we still need to run the standard
	// clean up and drop the caches if anything was deleted to ensure cache
	// consistency before terminating the process.
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if d == nil {
				// readdir(dir) failed, so just return the error
				log.Errorf("could not list directory '%s': %s", path, err)
				return err
			}

			// Returning SkipDir when handling a readdir error will skip the
			// remainder of the parent directory.
			// This has been fixed in go1.19rc1 https://github.com/golang/go/commit/460fd63cccd2f1d16fc4b1b761545b1649e14e28
			// Until then return nil, this will skip the current directory
			// but continue processing the remainder of the parent directory.
			log.Errorf("could not list directory '%s': %s", path, err)
			return nil
		}

		err = ctx.Err()
		if err != nil {
			// context has been cancelled, terminate walking the directory tree
			return err
		}

		// only care about directories and regular files, ignore anything special
		t := d.Type()
		switch {
		case t.IsDir():
			// Attempt to unlink the directory in case the directory is empty
			// from a previous cull.
			err := unix.Rmdir(path)
			if err == nil {
				// Directory was removed.
				return filepath.SkipDir
			} else {
				// Ignore any errors as the directory might still contain files,
				// or could be in use.
				return nil
			}

		case t.IsRegular():
			stat, err := lstat(path)
			if err != nil {
				// log the error and continue walking the directory tree
				log.Errorf("could not stat '%s': %s", path, err)
				return nil
			}

			totalFiles++
			lastAccessed := time.Unix(stat.Atim.Unix())
			if lastAccessed.Before(cutoff) {
				err := unix.Unlink(path)
				if err != nil {
					// Not updating oldestFile when a file fails to unlink to
					// avoid running another culling cycle almost immediately
					// just to remove a single file.
					// This assumes this error is rare and only happens
					// occasionally due to a file in use at the time of culling.
					log.Errorf("could not unlink '%s': %s", path, err)
				} else {
					culledSomething = true
				}
				filesCulled++
			} else if lastAccessed.Before(oldestFile) {
				oldestFile = lastAccessed
			}
		}

		return nil
	})

	filesRetained := totalFiles - filesCulled
	log.Infof("cull: %d files removed, %d files retained", filesCulled, filesRetained)
	// oldestFile will be logged later once the next cull check has been calculated

	if oldestFile == maxTime {
		// Either the cache is empty, or every file was culled. Set oldestFile
		// to the zero time to signal that there's no oldest file.
		oldestFile = time.Time{}
	}

	return culledSomething, oldestFile
}

// nextCullCheck calculates when to scan the cache for files to cull based upon
// the oldest file in the cache, and the quite period.
func (cfg *culler) nextCullCheck(oldestFile time.Time) time.Duration {
	if oldestFile.IsZero() {
		return cfg.quietPeriod
	}

	ageOfFile := time.Since(oldestFile)
	oldestFileExpires := cfg.lastAccess - ageOfFile

	if oldestFileExpires < cfg.quietPeriod {
		// Do not cull more frequently than the quiet period.
		return cfg.quietPeriod
	} else {
		// Wait until the oldest file in the cache can be deleted before
		// attempting another cull. Add on a little extra time to increase the
		// chance of culling multiple files and avoid scanning the entire
		// directory tree just to delete a single file.
		// The standard polling interval can be used as a good indicator of
		// how long to wait.
		return oldestFileExpires + cfg.interval
	}
}

func dropVMCache() {
	log.Info("dropping VM caches")

	f, err := os.OpenFile(dropCachesPath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		goto fail
	}

	_, err = f.Write([]byte{'2', '\n'})
	if err != nil {
		goto fail
	}

	err = f.Close()
	f = nil
	if err != nil {
		goto fail
	}

	return

fail:
	if f != nil {
		f.Close()
	}
	log.Errorf("could not drop VM caches: %v\n", err)
}

func statfs(path string) (unix.Statfs_t, error) {
	var s unix.Statfs_t
	for {
		err := unix.Statfs(path, &s)
		if err != unix.EINTR {
			return s, err
		}
	}
}

func lstat(path string) (unix.Stat_t, error) {
	var s unix.Stat_t
	for {
		err := unix.Lstat(path, &s)
		if err != unix.EINTR {
			return s, err
		}
	}
}

// Sleep pauses the current goroutine for at least the duration d or until the
// context is cancelled.
// A negative or zero duration causes Sleep to return immediately.
// Returns true if the sleep completed, or false if the context was cancelled.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}
