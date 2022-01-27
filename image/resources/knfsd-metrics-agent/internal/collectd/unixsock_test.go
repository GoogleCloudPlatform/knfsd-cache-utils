package collectd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"collectd.org/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testdir string

func value(v float64) *api.ValueList {
	return &api.ValueList{
		Identifier: api.Identifier{
			Host:   "localhost",
			Plugin: "test",
			Type:   "gauge",
		},
		Interval: 60 * time.Second,
		Values: []api.Value{
			api.Gauge(v),
		},
	}
}

func expected(v float64) string {
	return fmt.Sprintf("PUTVAL \"localhost/test/gauge\" interval=60.000 N:%.15g", v)
}

func TestMain(m *testing.M) {
	var err error
	testdir, err = os.MkdirTemp("", "")
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()

	err = os.RemoveAll(testdir)
	if err != nil {
		log.Println(err)
	}

	os.Exit(code)
}

func TestUnixSock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	path := filepath.Join(testdir, "collectd.sock")
	writer := newUnix(path)

	// Try to write while socket does not exist.
	// Initial connection attempt will fail.
	err := writer.Write(ctx, value(1))
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, initialDelay, writer.retry.delay)
	writer.retry.reset()

	// Listen on the unix socket
	l, err := listen(path)
	require.NoError(t, err)
	defer l.close()

	// Socket now listening, this attempt will succeed.
	err = writer.Write(ctx, value(2))
	require.NoError(t, err)
	assert.Equal(t, expected(2), l.next())

	// Stop the listener, but keep the socket file.
	err = l.stop()
	require.NoError(t, err)

	// The listener is now closed but the client connection is unaware until
	// it next attempts to write a value.
	err = writer.Write(ctx, value(3))
	assert.ErrorIs(t, err, syscall.EPIPE) // broken pipe
	assert.Equal(t, initialDelay, writer.retry.delay)
	writer.retry.reset()

	// Start the listener back up
	err = l.start()
	require.NoError(t, err)

	// Logging another value should trigger a reconnect attempt
	err = writer.Write(ctx, value(4))
	require.NoError(t, err)
	assert.Equal(t, expected(4), l.next())

	// Simulate logging an invalid metric
	l.respondWith("-1 Missing identifier and/or value-list.")
	err = writer.Write(ctx, value(5))
	require.ErrorIs(t, err, RemoteError("Missing identifier and/or value-list."))

	// Simulate remote listener restarting
	err = l.stop()
	require.NoError(t, err)

	err = l.start()
	require.NoError(t, err)

	err = writer.Write(ctx, value(6))
	require.NoError(t, err)
	assert.Equal(t, expected(6), l.next())
}

type socketListener struct {
	path         string
	nextResponse chan string
	receive      chan string
	listener     *net.UnixListener

	m    *sync.Mutex
	conn net.Conn
}

func listen(path string) (*socketListener, error) {
	l := &socketListener{
		path:         path,
		m:            &sync.Mutex{},
		nextResponse: make(chan string, 1),
	}
	err := l.start()
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *socketListener) start() error {
	var err error
	l.listener, err = net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: l.path})
	if err != nil {
		return err
	}

	l.receive = make(chan string, 1)
	go l.run()
	return nil
}

func (l *socketListener) stop() error {
	l.m.Lock()
	defer l.m.Unlock()

	err := l.listener.Close()
	l.listener = nil

	if l.conn != nil {
		e := l.conn.Close()
		if err == nil {
			err = e
		}
	}

	return err
}

func (l *socketListener) close() error {
	err := l.stop()
	e := os.Remove(l.path)
	if err == nil {
		err = e
	}
	return err
}

func (l *socketListener) run() {
	listener := l.listener
	receive := l.receive
	defer close(receive)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		l.m.Lock()
		stopped := listener != l.listener
		if stopped {
			conn.Close()
		} else {
			l.conn = conn
		}
		l.m.Unlock()

		if stopped {
			return
		}

		s := bufio.NewScanner(conn)
		for s.Scan() {
			select {
			case receive <- s.Text():
			default:
				panic("could not receive request")
			}

			var msg string
			select {
			case msg = <-l.nextResponse:
			default:
				msg = "0 Success: 1 value has been dispatched."
			}

			_, err = fmt.Fprintln(conn, msg)
			if err != nil {
				return
			}
		}
	}
}

func (l *socketListener) next() string {
	return <-l.receive
}

func (l *socketListener) respondWith(msg string) {
	select {
	case l.nextResponse <- msg:
	default:
		panic("could not queue next response")
	}
}
