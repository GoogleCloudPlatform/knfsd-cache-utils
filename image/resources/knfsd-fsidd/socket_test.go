package main

import (
	"context"
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
