package proxy

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestBridge(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Server echoes back what it receives
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.Copy(conn, conn)
	}()

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Bridge(ln.Addr().String(), stdinR, stdoutW)
	}()

	go func() {
		stdinW.Write([]byte("hello"))
		time.Sleep(50 * time.Millisecond)
		stdinW.Close()
	}()

	buf := make([]byte, 1024)
	n, _ := stdoutR.Read(buf)
	got := string(buf[:n])

	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestConnectWithRetry(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := ln.Addr().String()
	ln.Close()

	// Restart after 300ms
	go func() {
		time.Sleep(300 * time.Millisecond)
		newLn, _ := net.Listen("tcp", addr)
		if newLn != nil {
			conn, _ := newLn.Accept()
			if conn != nil {
				conn.Close()
			}
			newLn.Close()
		}
	}()

	conn, err := ConnectWithRetry(addr, 5*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestConnectWithRetryTimeout(t *testing.T) {
	_, err := ConnectWithRetry("127.0.0.1:59999", 300*time.Millisecond, 100*time.Millisecond)
	if err == nil {
		t.Error("should error on timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should mention timeout, got: %s", err)
	}
}
