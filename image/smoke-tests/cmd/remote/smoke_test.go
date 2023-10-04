package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

func TestSmoke(t *testing.T) {
	// These tests need to be executed in order as some steps are dependent on
	// previous steps.
	require.True(t, t.Run("prepare", func(t *testing.T) {
		err := Sudo(`
		set -xe

		# Unmount the shares if they're already mounted, this allows re-running
		# the smoke tests.
		if mountpoint -q /mnt/source; then
			umount -f /mnt/source
		fi

		if mountpoint -q /mnt/proxy; then
			umount -f /mnt/proxy
		fi

		# Create the directories for the mounts.
		# Mark them as immutable so that if the mount fails no data
		# can be accidentally wrote to the local disk.
		mkdir -p /mnt/source /mnt/proxy
		chattr +i /mnt/source /mnt/proxy

		# Add symlinks to the test directories to simplify the tests.
		mkdir -p /test
		`)
		require.NoError(t, err)
	}))

	require.True(t, t.Run("mount source", func(t *testing.T) {
		// Mount the source directly so that we can bypass the proxy to verify
		// that data was wrote to the source, and setting up data to be read by
		// the proxy.
		sourceMount := fmt.Sprintf("%s:/files", sourceHost)
		err := Sudo(fmt.Sprintf(`
		set -e

		# Disable as much caching on the client as possible so that read/writes
		# always go back to the source.
		mount "%s" /mnt/source -o vers=3,proto=tcp,noac,noatime,nocto,rsize=1048576,wsize=1048576,lookupcache=none,nolock

		# Create the test directory if it doesn't already exist and
		# assign ownership to our test user.
		mkdir -p "/mnt/source/smoke-tests"
		rm -f "/mnt/source/smoke-tests/*"
		chown "$SUDO_UID:$SUDO_GID" "/mnt/source/smoke-tests"
		chmod 775 "/mnt/source/smoke-tests"

		ln -snf /mnt/source/smoke-tests /test/source
		`, sourceMount))
		require.NoError(t, err)
	}))

	require.True(t, t.Run("mount proxy", func(t *testing.T) {
		proxyMount := fmt.Sprintf("%s:/files", proxyHost)
		err := Sudo(fmt.Sprintf(`
		set -e

		# Disable as much caching on the client as possible so that the client
		# client has to re-query the proxy during the smoke tests.
		mount "%s" /mnt/proxy -o vers=3,proto=tcp,noac,noatime,nocto,rsize=1048576,wsize=1048576,lookupcache=none,nolock

		ln -snf /mnt/proxy/smoke-tests /test/proxy
		`, proxyMount))
		require.NoError(t, err)
	}))

	t.Run("read via proxy", func(t *testing.T) {
		t.Parallel()

		name, err := createRandomFile("/test/source", "read.*")
		require.NoError(t, err)
		t.Cleanup(func() { removeTestFile(name) })

		err = writeRandomData("/test/source/"+name, 1*MB)
		require.NoError(t, err)

		assertFilesEqual(t, "/test/source/"+name, "/test/proxy/"+name)
	})

	t.Run("write via proxy", func(t *testing.T) {
		t.Parallel()

		name, err := createRandomFile("/test/proxy", "write.*")
		require.NoError(t, err)
		t.Cleanup(func() { removeTestFile(name) })

		err = writeRandomData("/test/proxy/"+name, 1*MB)
		require.NoError(t, err)

		assertFilesEqual(t, "/test/proxy/"+name, "/test/source/"+name)
	})

	t.Run("metadata caches positive lookups", func(t *testing.T) {
		t.Parallel()

		// Create a random file on the source and stat it via the proxy. Then
		// remove the file from the source. The proxy should still have the
		// metadata cached as the proxy will be unaware the file was removed.

		name, err := createRandomFile("/test/source", "meta.*")
		require.NoError(t, err)
		t.Cleanup(func() { removeTestFile(name) })

		before, err := os.Stat("/test/proxy/" + name)
		require.NoError(t, err)

		// A single stat doesn't always cache the metadata, so re-read.
		for i := 0; i < 100; i++ {
			_, err = os.Stat("/test/proxy/" + name)
			require.NoError(t, err)
		}

		// Remove the file directly via the source so the proxy is unaware.
		err = os.Remove("/test/source/" + name)
		require.NoError(t, err)

		// Drop this machines caches so that it has to go back to the proxy.
		err = dropLocalVMCaches()
		require.NoError(t, err)

		// Read the metadata from the proxy for the now non-existent file.
		require.FileExists(t, "/test/proxy/"+name)
		assert.NoFileExists(t, "/test/source/"+name)

		after, err := os.Stat("/test/proxy/" + name)
		require.NoError(t, err)
		assert.Equal(t, before, after)
	})

	t.Run("metadata caches negative lookups", func(t *testing.T) {
		t.Parallel()

		// Similar to the positive test, only we're going to stat a file that
		// doesn't exist, then create it. The proxy should continue to think the
		// file doesn't exist.

		// Grab a random file name.
		name, err := createRandomFile("/test/source", "meta.*")
		require.NoError(t, err)
		t.Cleanup(func() { removeTestFile(name) })

		// Remove the file and ensure it doesn't exist according to the proxy.
		err = os.Remove("/test/source/" + name)
		require.NoError(t, err)
		require.NoFileExists(t, "/test/source/"+name)
		require.NoFileExists(t, "/test/proxy/"+name)

		// Ensure the negative lookup is cached on the proxy.
		for i := 0; i < 100; i++ {
			os.Stat("/test/proxy/" + name)
		}

		// Create the file directly on the source so the proxy is unaware.
		f, err := os.Create("/test/source/" + name)
		require.NoError(t, err)
		f.Close()

		// Drop this machines caches so that it has to go back to the proxy.
		err = dropLocalVMCaches()
		require.NoError(t, err)

		assert.NoFileExists(t, "/test/proxy/"+name)
		assert.FileExists(t, "/test/source/"+name)
	})

	t.Run("proxy caches file data", func(t *testing.T) {
		// This test checks two properties of the cache:
		// * The cache actually caches the file data
		// * The file data is cached on disk using cachefilesd
		// It is not worth separating these two tests out as the setup for them
		// is identical, and working with a 1GB file makes the test
		// comparatively slow.
		//
		// Not using a tool such as fincore as we're explicitly testing that the
		// file data was cached to disk using FS-Cache (cachefilesd).
		//
		// If the cache is full, this test would fail as old data would need to
		// be evicted to store the new data, resulting in a net difference of
		// zero. However, this is not considered an issue as these smoke tests
		// are designed to be run on a new instance of the proxy that was
		// created specifically for running the smoke tests. As such its only
		// likely this will happen when developing the smoke tests where a
		// developer is likely to keep re-running the same tests on the same
		// instance.

		// This could fail if the cache is full, as old data will be evicted
		// to make space for the new data. However, in practice this is
		// unlikely when running smoke tests as the cache size is 350 GB.
		initialSize, err := fsCacheSize()
		require.NoError(t, err)

		name, err := createRandomFile("/tmp", "large.*")
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = os.Remove("/tmp/" + name)
			removeTestFile(name)
		})

		// Seed large (1G) file for the cache test. Create the file locally so
		// that it can be compared after deleting the source.
		err = writeRandomData("/tmp/"+name, 1*GB)
		require.NoError(t, err)

		err = copyFile("/tmp/"+name, "/test/source/"+name)
		require.NoError(t, err)

		// Read the file through the proxy and ensure it matches.
		assertFilesEqual(t, "/tmp/"+name, "/test/proxy/"+name)

		// Read it a few more times to ensure it is fully cached
		for i := 0; i < 10; i++ {
			err = copyFile("/test/proxy/"+name, "/dev/null")
			require.NoError(t, err)
		}

		err = dropLocalVMCaches()
		require.NoError(t, err)

		// Remove the file from the source, then test that the proxy still
		// serves the file from the cache.
		err = os.Remove("/test/source/" + name)
		require.NoError(t, err)

		assertFilesEqual(t, "/tmp/"+name, "/test/proxy/"+name)

		cacheSize, err := fsCacheSize()
		require.NoError(t, err)

		initialSize = initialSize / MB
		cacheSize = cacheSize / MB
		diff := cacheSize - initialSize
		if diff < 970 || diff > 1080 {
			msg := fmt.Sprintf(
				"Expected cache delta to be between 970M and 1080M, but was %d MB\n"+
					"Initial size: %d MB\n"+
					"  Final size: %d MB",
				diff, initialSize, cacheSize,
			)
			t.Error(msg)
		}
	})
}

func createRandomFile(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	f.Close()
	return filepath.Base(f.Name()), nil
}

func writeRandomData(path string, size uint64) error {
	// Use dd as it's already solved the hard problems of setting cache flags
	// and efficiently copying data.
	_, err := exec.Command("dd",
		"if=/dev/urandom",
		"of="+path,
		"bs=4M",
		"count="+strconv.FormatUint(size, 10),
		"iflag=count_bytes",
		"oflag=nocache",
		"status=none",
	).Output()
	return err
}

func assertFilesEqual(t *testing.T, expected, actual string) {
	t.Helper()
	out, err := exec.Command("cmp", expected, actual).CombinedOutput()
	if err != nil {
		t.Logf("files were different: %s", out)
	}
}

func removeTestFile(name string) {
	// Do our best to remove the file, try via the proxy first so that
	// the proxy is aware the file was removed.
	_ = os.Remove("/test/proxy/" + name)
	_ = os.Remove("/test/source/" + name)
}

func dropLocalVMCaches() error {
	unix.Sync()
	return os.WriteFile("/proc/sys/vm/drop_caches", []byte("3\n"), 0)
}

func fsCacheSize() (uint64, error) {
	u, err := proxy.CacheUsage()
	if err != nil {
		return 0, err
	}
	return u.BytesUsed, nil
}

func copyFile(src, dst string) error {
	// It's easier to just fork out to cp than try to re-implement the logic.
	cmd := exec.Command("cp", "--", src, dst)
	_, err := cmd.Output()
	return err
}
