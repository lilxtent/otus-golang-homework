package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type operation string

const (
	operationSend    operation = "send"
	operationReceive operation = "receive"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "invalid amount of args, expected: go-telnet [--timeout=10s] host port")
		os.Exit(2)
	}

	address := net.JoinHostPort(flag.Arg(0), flag.Arg(1))
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT)
	defer signal.Stop(signalCh)

	type result struct {
		operation operation
		err       error
	}

	errCh := make(chan result, 2)
	go func() {
		errCh <- result{operation: operationSend, err: client.Send()}
	}()
	go func() {
		errCh <- result{operation: operationReceive, err: client.Receive()}
	}()

	select {
	case <-signalCh:
		_, _ = fmt.Fprintln(os.Stderr, "\n...SIGINT")
	case res := <-errCh:
		switch {
		case res.err != nil:
			_, _ = fmt.Fprintln(os.Stderr, res.err)
		case res.operation == operationSend:
			_, _ = fmt.Fprintln(os.Stderr, "...EOF")
		default:
			_, _ = fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	}

	if err := client.Close(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
