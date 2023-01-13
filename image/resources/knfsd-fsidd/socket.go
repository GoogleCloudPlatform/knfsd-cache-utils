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
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"

	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

var (
	ErrUnknownCommand  = errors.New("unknown command")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrServerClosed    = net.ErrClosed
)

const PacketMaxLength = unix.PathMax * 2

type Handler func(ctx context.Context, arg string) (string, error)

type server struct {
	listener *net.UnixListener
	handlers map[string]Handler

	connectionID atomic.Uint64
	connections  map[*connection]struct{}

	mu            sync.Mutex
	c             sync.Cond
	listenerGroup sync.WaitGroup
	inShutdown    atomic.Bool
}

func newServer(socketPath string) (*server, error) {
	addr := &net.UnixAddr{Net: "unixpacket", Name: socketPath}
	l, err := net.ListenUnix(addr.Network(), addr)
	if err != nil {
		return nil, err
	}
	return &server{
		listener: l,
		handlers: make(map[string]Handler),
	}, nil
}

// Handle registers the handler for a given command. If a handler already exists
// for a pattern, Handle panics. `cmd` is case-insensitive.
func (s *server) Handle(cmd string, handler Handler) {
	cmd = strings.ToUpper(cmd)
	if cmd == "" {
		panic("cmd required")
	}
	if handler == nil {
		panic("nil handler")
	}
	if _, duplicate := s.handlers[cmd]; duplicate {
		panic("multiple registrations for " + cmd)
	}
	s.handlers[cmd] = handler
}

// Serve accepts incoming connections.
//
// Serve always returns a non-nil error. After Shutdown or Close the returned
// error is ErrServerClosed.
func (s *server) Serve() error {
	if !s.trackListener(true) {
		return ErrServerClosed
	}
	defer s.trackListener(false)

	for {
		rw, err := s.listener.AcceptUnix()
		if err != nil {
			return err
		}
		c := &connection{
			id:       s.connectionID.Add(1),
			s:        s,
			rw:       rw,
			handlers: s.handlers,
		}

		s.trackConnection(c, true)
		go c.Serve()
	}
}

func (s *server) trackListener(add bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		if s.shuttingDown() {
			return false
		}
		s.listenerGroup.Add(1)
	} else {
		s.listenerGroup.Done()
	}
	return true
}

func (s *server) trackConnection(c *connection, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections == nil {
		s.connections = make(map[*connection]struct{})
	}
	if add {
		s.connections[c] = struct{}{}
	} else {
		delete(s.connections, c)
		if len(s.connections) == 0 && s.shuttingDown() {
			// When the last connection is closed signal the condition in case
			// shutdown is waiting.
			s.c.Signal()
		}
	}
}

func (s *server) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (s *server) Close() error {
	return s.stop()
}

func (s *server) Shutdown(ctx context.Context) error {
	err := s.stop()

	// Create our own child context so that we can cancel it when Shutdown
	// returns. This prevents leaking the goroutine if the parent context is
	// not cancelled.
	// When the context is cancelled send a signal to wake the waiting thread.
	// Not worrying about race-conditions here with multiple calls to Signal as
	// it's only expected that Shutdown will be called once.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		s.c.Signal()
	}()

	s.mu.Lock()
	defer s.mu.Lock()

	// If there are any connections, wait until they have terminated, or until
	// ctx is cancelled.
	if len(s.connections) > 0 {
		s.c.Wait()
	}

	return err
}

func (s *server) stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	// Stop accepting new connections.
	s.inShutdown.Store(true)
	err = multierr.Append(err, s.listener.Close())

	// Wait for the listener to shutdown to avoid race between accept and
	// tracking the new connection.
	s.mu.Unlock()
	s.listenerGroup.Wait()
	s.mu.Lock()

	for c := range s.connections {
		err = multierr.Append(err, c.Close())
	}

	return err
}

type connection struct {
	id         uint64
	s          *server
	rw         *net.UnixConn
	handlers   map[string]Handler
	inShutdown atomic.Bool
	cancel     context.CancelFunc
}

func (c *connection) shuttingDown() bool {
	return c.inShutdown.Load()
}

func (c *connection) Serve() {
	defer func() {
		err := c.Close()
		c.s.trackConnection(c, false)
		if err != nil && !isClosed(err) {
			log.Warn.Printf("[%d] %s", c.id, err)
		}
		log.Debug.Printf("[%d] connection closed", c.id)
	}()

	log.Debug.Printf("[%d] received connection", c.id)

	// Make the buffer 1 byte larger than the max allowed packet length to
	// detect truncated packets. If we read > PacketMaxLength bytes then the
	// packet would have been truncated if the buffer was PacketMaxLength.
	buf := make([]byte, PacketMaxLength+1)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithID(ctx, c.id)
	c.cancel = cancel

	// Check inShutdown as r.Scan could have another command buffered.
	// There is a race condition between setting setting inShutdown, closing
	// the connection and scan returning the next command but that's ok, as
	// the command will have the grace period to complete its work before being
	// terminated.
	for !c.shuttingDown() {
		n, err := c.rw.Read(buf)
		if err != nil {
			switch {
			case isClosed(err):
				// connection closed
				log.Debug.Printf("[%d] received EOF", c.id)
				return

			default:
				// Non-recoverable error, reset the connection.
				log.Error.Printf("[%d] read error: %s", c.id, err)
				return
			}
		}

		if n > PacketMaxLength {
			// message was truncated, return an error to the client
			log.Warn.Printf("[%d] message truncated, ignoring", c.id)
			c.writeError("message truncated")
			continue
		}

		line := string(buf[0:n])
		log.Debug.Printf("[%d] => %q", c.id, line)

		cmd, arg, _ := cut(line, " ")
		cmd = strings.ToUpper(cmd)

		err = c.execute(ctx, cmd, arg)
		if err != nil {
			log.Error.Printf("[%d] error executing command: %s", c.id, err)
			// Non-recoverable error, reset the connection.
			return
		}
	}
}

func (c *connection) execute(ctx context.Context, cmd, arg string) error {
	h := c.handlers[cmd]
	if h == nil {
		return c.writeError(fmt.Sprintf("unknown command %q", cmd))
	}

	response, err := h(ctx, arg)
	if err != nil {
		// TODO: figure out if error is recoverable, for now assume it is
		return c.writeError(err.Error())
	}

	return c.write("+ " + response)
}

func (c *connection) writeError(msg string) error {
	return c.write("- " + msg)
}

func (c *connection) write(msg string) error {
	log.Debug.Printf("[%d] <= %q", c.id, msg)

	b := []byte(msg)
	nr, err := c.rw.Write(b)
	if err != nil {
		return err
	}
	if nr < len(b) {
		// This shouldn't happen as net.Conn#Write should only have a short
		// write if there's some sort of connection error; but just in case
		// consider this an un-recoverable connection error.
		return io.ErrShortWrite
	}
	return nil
}

func (c *connection) Shutdown() error {
	// Close the read side of the connection to stop reading any more commands,
	// but do not interrupt any command already in progress.
	// Keep the write open to allow for a final response.
	log.Debug.Printf("[%d] shutdown", c.id)
	c.inShutdown.Store(true)
	return c.rw.CloseRead()
}

func (c *connection) Close() error {
	// Close the connection and cancel any commands that is in progress.
	// Normally commands complete quickly, so it's most likely any command in
	// progress is stuck in a retry loop.
	log.Debug.Printf("[%d] close", c.id)
	c.inShutdown.Store(true)
	err := c.rw.CloseRead()
	if c.cancel != nil {
		c.cancel()
	}
	return err
}

func dial(socketPath string) (*net.UnixConn, error) {
	addr := &net.UnixAddr{Net: "unixpacket", Name: socketPath}
	c, err := net.DialUnix(addr.Network(), nil, addr)
	if err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func isClosed(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)
}

func cut(s, sep string) (before string, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	} else {
		return s, "", false
	}
}
