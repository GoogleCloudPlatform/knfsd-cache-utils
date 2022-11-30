package main

import (
	"context"
	"errors"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"
	"github.com/googleapis/gax-go/v2"
	"github.com/jackc/pgconn"
)

func withRetry(ctx context.Context, fn func() error) error {
	deadline := time.Now().Add(5 * time.Minute)
	backoff := &gax.Backoff{
		// Because of jitter, this will pick a time between 1ns and Interval
		// Keeping the initial interval small as the kernel is waiting for this
		// response. Though not too small otherwise it's likely to cause another
		// collision and have to retry again.
		Initial:    50 * time.Millisecond,
		Max:        60 * time.Second,
		Multiplier: 2,
	}

	// Keep retrying until the deadline is reached.
	var err error
	for {
		// Not interrupting an attempt once it has started, the deadline only
		// applies to the sleep/retry loop.
		err = fn()
		if err == nil {
			return nil
		}
		if !ShouldRetry(err) {
			return err
		}

		pause := backoff.Pause()
		// Check there's enough time remaining before the deadline for another attempt.
		if !time.Now().Add(pause).Before(deadline) {
			return err
		}

		log.Debug.Printf("[%d] RETRY (%s): %v", log.ID(ctx), pause, err)

		// Allow sleep to be interrupted by the context being cancelled to allow
		// for graceful shutdown.
		err = sleep(ctx, pause)
		if err != nil {
			return err
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	if d < 1 {
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var retryableErrorCodes = map[string]struct{}{
	"08006": {}, // connection_failure
	"08001": {}, // sqlclient_unable_to_establish_sqlconnection
	"08004": {}, // sqlserver_rejected_establishment_of_sqlconnection
	"08P01": {}, // protocol_violation

	// constraint errors
	"23505": {}, // unique_violation

	// transaction errors, transaction will be rolled back so try again
	"40000": {}, // transaction_rollback
	"40002": {}, // transaction_integrity_constraint_violation
	"40001": {}, // serialization_failure
	"40003": {}, // statement_completion_unknown
	"40P01": {}, // deadlock_detected

	// consider a longer delay for these errors
	// the server might recover itself as other clients disconnect, or if the
	// database auto-grows.
	"53000": {}, // insufficient_resources
	"53100": {}, // disk_full
	"53200": {}, // out_of_memory
	"53300": {}, // too_many_connections
}

func ShouldRetry(err error) bool {
	if pgconn.SafeToRetry(err) {
		return true
	}
	if pgconn.Timeout(err) {
		return true
	}

	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		_, retry := retryableErrorCodes[pgerr.Code]
		return retry
	}
	return false
}
