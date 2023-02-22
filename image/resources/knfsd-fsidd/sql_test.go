//go:build test.sql

package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDatabaseURL string = os.Getenv("TEST_DATABASE_URL")
var errTestDatabaseNotSet = errors.New("TEST_DATABASE_URL not set, cannot run SQL tests")

// connect creates a new connection to the test postgresql instance.
func connectTest() (FSIDSource, error) {
	ctx := context.Background()

	if testDatabaseURL == "" {
		return FSIDSource{}, errTestDatabaseNotSet
	}

	pgConfig, err := pgxpool.ParseConfig(testDatabaseURL)
	if err != nil {
		return FSIDSource{}, err
	}

	pool, err := pgxpool.ConnectConfig(ctx, pgConfig)
	if err != nil {
		return FSIDSource{}, err
	}

	source := FSIDSource{db: pool, tableName: "fsid-test"}

	// clean up after previous test
	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS \"fsid-test\"")
	if err != nil {
		goto fail
	}

	// start with a fresh table and sequence for every test
	err = source.CreateTable(ctx)
	if err != nil {
		goto fail
	}

	return source, nil

fail:
	pool.Close()
	return FSIDSource{}, err
}

func TestFSIDSource(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		source, err := connectTest()
		require.NoError(t, err)
		defer source.db.Close()

		ctx := context.Background()
		allocated_fsid, err := source.AllocateFSID(ctx, "/foo")
		require.NoError(t, err)
		assert.NotEqual(t, int32(0), allocated_fsid)

		fsid, err := source.GetFSID(ctx, "/foo")
		if assert.NoError(t, err) {
			assert.Equal(t, allocated_fsid, fsid)
		}

		path, err := source.GetPath(ctx, allocated_fsid)
		if assert.NoError(t, err) {
			assert.Equal(t, "/foo", path)
		}
	})

	t.Run("MissingFSID", func(t *testing.T) {
		source, err := connectTest()
		require.NoError(t, err)
		defer source.db.Close()

		path, err := source.GetPath(context.Background(), 1)
		assert.Equal(t, "", path)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("MissingPath", func(t *testing.T) {
		source, err := connectTest()
		require.NoError(t, err)
		defer source.db.Close()

		fsid, err := source.GetFSID(context.Background(), "/foo")
		assert.Equal(t, int32(0), fsid)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestAllocateFSID(t *testing.T) {
	source, err := connectTest()
	require.NoError(t, err)
	defer source.db.Close()

	t.Run("IsConflict", func(t *testing.T) {
		ctx := context.Background()
		_, err = source.AllocateFSID(ctx, "/foo")
		require.NoError(t, err)

		_, err = source.AllocateFSID(ctx, "/foo")
		require.Error(t, err)
		require.True(t, IsConflict(err))
		require.True(t, ShouldRetry(err))
	})

	// To run this test multiple times use:
	//   ./test.sh -run TestAllocateFSID/Race -count=X
	t.Run("Race", func(t *testing.T) {
		// Try to race multiple workers against each other all calling
		// AllocateFSID at the same time. Only one worker should be successful,
		// the remainder should all get a unique constraint violation.
		const worker_count = 10

		start := sync.WaitGroup{}
		done := sync.WaitGroup{}
		start.Add(1)
		done.Add(worker_count)

		path := uuid.NewString()
		worker := func(err *error) {
			start.Wait()
			_, *err = source.AllocateFSID(context.Background(), path)
			done.Done()
		}

		errors := make([]error, worker_count)
		for i := 0; i < worker_count; i++ {
			go worker(&errors[i])
		}
		start.Done()
		done.Wait()

		// Check that all the errors were conflicts
		var unexpected []error
		for _, err := range errors {
			if err != nil && !IsConflict(err) {
				unexpected = append(unexpected, err)
			}
		}
		require.Empty(t, unexpected)

		error_count := 0
		for _, e := range errors {
			if e != nil {
				error_count++
				err = e
			}
		}
		// Only one worker should succeed
		require.Equal(t, worker_count-1, error_count)
	})
}
