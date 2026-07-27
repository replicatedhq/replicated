package credentials

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartLocalWebServerRespectsCancel(t *testing.T) {
	port, err := getAvailablePort()
	if err != nil {
		t.Fatalf("getAvailablePort: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := startLocalWebServer(ctx, port)
		errCh <- err
	}()

	// Wait until the server is accepting connections.
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 50*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("server did not start listening: %v", dialErr)
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for startLocalWebServer to return after cancel")
	}

	// Port should be free again after Shutdown.
	deadline = time.Now().Add(2 * time.Second)
	for {
		ln, listenErr := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if listenErr == nil {
			_ = ln.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("port %d still in use after cancel/shutdown: %v", port, listenErr)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestStartLocalWebServerSourceUsesShutdown(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("fetch.go")
	if err != nil {
		t.Fatalf("read fetch.go: %v", err)
	}
	content := string(src)
	if !strings.Contains(content, "server.Shutdown(shutdownCtx)") {
		t.Fatal("startLocalWebServer must call server.Shutdown")
	}
	if strings.Contains(content, "ctx.Done()\n\t\treturn") {
		// Old broken pattern called Done() without selecting on it for cancel.
		// We only check Shutdown presence as the structural requirement.
	}
	if !strings.Contains(content, "case <-ctx.Done():") {
		t.Fatal("startLocalWebServer must select on ctx.Done for cancellation")
	}
}
