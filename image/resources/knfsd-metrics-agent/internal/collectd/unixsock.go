package collectd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"collectd.org/api"
	"collectd.org/format"
)

var (
	initialDelay    = 5 * time.Second
	maxDelay        = 10 * time.Minute
	ErrRetryBackoff = errors.New("waiting to retry")
)

type RemoteError string

func (err RemoteError) Error() string {
	if err == "" {
		return "unknown error from collectd"
	} else {
		return string(err)
	}
}

type unix struct {
	path string

	conn   net.Conn
	writer api.Writer
	reader *bufio.Scanner

	retry backoff
	m     *sync.Mutex
}

type backoff struct {
	delay time.Duration
	next  time.Time
}

func (b *backoff) increment() {
	if b.delay == 0 {
		b.delay = initialDelay
	} else {
		b.delay *= 2
		if b.delay > maxDelay {
			b.delay = maxDelay
		}
	}
	b.next = time.Now().Add(b.delay)
}

func (b *backoff) reset() {
	b.delay = 0
	b.next = time.Time{}
}

func (b *backoff) ready() bool {
	return b.delay == 0 || time.Now().After(b.next)
}

func newUnix(path string) *unix {
	return &unix{path: path, m: &sync.Mutex{}}
}

func (u *unix) connect() error {
	if u.conn != nil {
		return nil
	}
	if !u.retry.ready() {
		return ErrRetryBackoff
	}

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(context.TODO(), "unix", u.path)
	if err != nil {
		u.retry.increment()
		return err
	}

	u.conn = conn
	u.writer = format.NewPutval(conn)
	u.reader = bufio.NewScanner(conn)
	u.retry.reset()
	return nil
}

func (u *unix) Write(ctx context.Context, vl *api.ValueList) error {
	u.m.Lock()
	defer u.m.Unlock()

	retry, err := u.write(ctx, vl)
	if retry {
		recovered := u.retryWrite(ctx, vl)
		if recovered {
			return nil
		}
	}

	return err
}

func (u *unix) retryWrite(ctx context.Context, vl *api.ValueList) bool {
	_, err := u.write(ctx, vl)
	return err == nil
}

func (u *unix) write(ctx context.Context, vl *api.ValueList) (bool, error) {
	err := u.connect()
	if err != nil {
		return false, err
	}

	err = u.writer.Write(ctx, vl)
	if err != nil {
		if isNetworkError(err) {
			// can reconnect and try again
			u.close()
			return true, err
		}

		return false, err
	}

	err = u.read()
	return false, err
}

func (u *unix) read() error {
	if !u.reader.Scan() {
		err := u.reader.Err()
		u.close()
		return err
	}

	// Status line in the format "<code> <msg>"
	// e.g. -1 Missing identifier and/or value-list.
	line := u.reader.Text()
	split := strings.IndexRune(line, ' ')
	if split < 1 {
		u.close()
		return fmt.Errorf("could not parse collectd response: %w", RemoteError(line))
	}

	msg := line[split+1:]
	code, err := strconv.Atoi(line[0:split])
	if err != nil {
		u.close()
		return fmt.Errorf("could not parse collectd response: %w", RemoteError(line))
	}

	if code < 0 {
		return RemoteError(msg)
	} else {
		return nil
	}
}

func (u *unix) close() {
	if u.conn != nil {
		err := u.conn.Close()
		if err != nil {
			log.Printf("WARN: could not close unix socket: %s\n", err)
		}
	}

	u.conn = nil
	u.writer = nil
	u.reader = nil
}

func isNetworkError(err error) bool {
	var op *net.OpError
	if errors.As(err, &op) {
		return true
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	if errors.Is(err, syscall.EPIPE) {
		return true
	}

	return false
}
