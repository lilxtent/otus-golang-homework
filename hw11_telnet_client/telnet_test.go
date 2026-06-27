package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := new(net.ListenConfig).Listen(context.Background(), "tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connect returns error", func(t *testing.T) {
		client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		require.Error(t, client.Connect())
	})

	t.Run("send multiline input", func(t *testing.T) {
		l, err := new(net.ListenConfig).Listen(context.Background(), "tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		input := "hello\nfrom\ntelnet\n"
		receivedCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			conn, err := l.Accept()
			if err != nil {
				errCh <- err
				return
			}
			defer func() { _ = conn.Close() }()

			request := make([]byte, len(input))
			if _, err = io.ReadFull(conn, request); err != nil {
				errCh <- err
				return
			}

			receivedCh <- string(request)
			errCh <- nil
		}()

		client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(bytes.NewBufferString(input)), &bytes.Buffer{})
		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		require.NoError(t, client.Send())
		require.NoError(t, <-errCh)
		require.Equal(t, input, <-receivedCh)
	})

	t.Run("receive multiple chunks", func(t *testing.T) {
		l, err := new(net.ListenConfig).Listen(context.Background(), "tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		errCh := make(chan error, 1)
		go func() {
			conn, err := l.Accept()
			if err != nil {
				errCh <- err
				return
			}
			defer func() { _ = conn.Close() }()

			for _, chunk := range []string{"hello\n", "from\n", "server\n"} {
				if _, err = conn.Write([]byte(chunk)); err != nil {
					errCh <- err
					return
				}
			}

			errCh <- nil
		}()

		out := &bytes.Buffer{}
		client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(&bytes.Buffer{}), out)
		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		require.NoError(t, client.Receive())
		require.NoError(t, <-errCh)
		require.Equal(t, "hello\nfrom\nserver\n", out.String())
	})

	t.Run("close before connect", func(t *testing.T) {
		client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		require.NoError(t, client.Close())
	})
}
