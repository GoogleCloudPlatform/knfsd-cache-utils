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
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	s, err := newServer("")
	require.NoError(t, err, "newServer(...)")
	defer s.Close()

	s.Handle("TEST", func(ctx context.Context, arg string) (string, error) {
		if arg == "" {
			return "", ErrInvalidArgument
		}
		return "Hello " + arg, nil
	})

	done := make(chan struct{})
	go func() {
		err := s.Serve()
		assert.ErrorIs(t, err, ErrServerClosed, "s.Serve()")
		close(done)
	}()

	buf := make([]byte, PacketMaxLength)
	c, err := dial(s.listener.Addr().String())
	require.NoError(t, err, "net.Dial(...)")
	defer c.Close()

	execute := func(t *testing.T, msg string) string {
		_, err = c.Write([]byte(msg))
		require.NoError(t, err, "c.Write(%q)", msg)
		n, err := c.Read(buf)
		require.NoError(t, err)
		return string(buf[0:n])
	}

	t.Run("valid command", func(t *testing.T) {
		result := execute(t, "TEST World")
		assert.Equal(t, "+ Hello World", result)
	})

	t.Run("invalid argument", func(t *testing.T) {
		result := execute(t, "TEST")
		assert.Equal(t, "- invalid argument", result)
	})

	t.Run("invalid command", func(t *testing.T) {
		result := execute(t, "UNKNOWN")
		assert.Equal(t, "- unknown command \"UNKNOWN\"", result)
	})

	err = s.Close()
	assert.NoError(t, err, "s.Close()")
	<-done
}

func TestNewServerFromFile(t *testing.T) {
	l, err := net.Listen("unixpacket", "")
	require.NoError(t, err)
	defer l.Close()

	f, err := l.(*net.UnixListener).File()
	require.NoError(t, err)
	defer f.Close()

	s, err := newServerFromFile(f)
	require.NoError(t, err)
	defer s.Close()
	go s.Serve()

	c, err := dial(l.Addr().String())
	require.NoError(t, err)
	defer c.Close()

	s.Handle("TEST", func(ctx context.Context, arg string) (string, error) {
		return arg, nil
	})
	_, err = c.Write([]byte("TEST 123"))
	require.NoError(t, err)

	buf := make([]byte, 20)
	n, err := c.Read(buf)
	require.NoError(t, err)

	assert.Equal(t, "+ 123", string(buf[0:n]))
}
