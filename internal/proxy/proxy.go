package proxy

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

func ConnectWithRetry(addr string, timeout time.Duration, interval time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, interval)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out connecting to %s", addr)
		}
		time.Sleep(interval)
	}
}

func Bridge(addr string, stdin io.Reader, stdout io.Writer) error {
	conn, err := ConnectWithRetry(addr, 5*time.Second, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// stdin -> TCP
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(conn, stdin)
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		errCh <- err
	}()

	// TCP -> stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(stdout, conn)
		errCh <- err
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
