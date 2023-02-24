package main

import (
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestShouldRetry(t *testing.T) {
	t.Run("ErrNoRows", func(t *testing.T) {
		// ErrNoRows should not be retried, as this is used by get_fsid and
		// get_path to signal that there's no record to return.
		assert.False(t, ShouldRetry(pgx.ErrNoRows))
	})
}
