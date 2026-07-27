package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
)

var (
	ErrNoBrowser = errors.New("no browser could be found")
)

func Fetch(endpoint string) error {
	// create a local web server with a single page
	// vendor portal will redirect to this page with a token

	localPort, err := getAvailablePort()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fullUri := fmt.Sprintf("%s/cli-login?redirect_uri=%s:%d/callback", endpoint, "http://localhost", localPort)
	if err := open.Start(fullUri); err != nil {
		if strings.Contains(err.Error(), "executable file not found in $PATH") {
			return ErrNoBrowser
		}
		return err
	}

	token, err := startLocalWebServer(ctx, localPort)
	if err != nil {
		return err
	}

	if err := SetCurrentCredentials(token); err != nil {
		return err
	}

	return nil
}

// startLocalWebServer handles the token redirect, returning the token
func startLocalWebServer(ctx context.Context, port int) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	errChan := make(chan error, 1)
	tokenChan := make(chan string, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := r.URL.Query().Get("nonce")
		if nonce == "" {
			select {
			case errChan <- fmt.Errorf("no nonce found in response"):
			default:
			}
			http.Error(w, "missing nonce", http.StatusBadRequest)
			return
		}

		exchange := r.URL.Query().Get("exchange")
		if exchange == "" {
			select {
			case errChan <- fmt.Errorf("no exchange found in response"):
			default:
			}
			http.Error(w, "missing exchange", http.StatusBadRequest)
			return
		}

		token, err := exchangeNonceForToken(exchange, nonce)
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
			http.Error(w, "token exchange failed", http.StatusBadGateway)
			return
		}

		select {
		case tokenChan <- token:
		default:
		}

		fmt.Fprintln(w, "Authentication successful. You may close this window.")
	})

	sm := http.NewServeMux()
	sm.Handle("/callback", handler)
	server := &http.Server{
		Addr:              net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		Handler:           sm,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	// Always stop admission and drain in-flight callback work when we leave.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		return "", fmt.Errorf("authentication timeout")
	case <-ctx.Done():
		return "", ctx.Err()
	case token := <-tokenChan:
		return token, nil
	case e := <-errChan:
		return "", e
	}
}

func exchangeNonceForToken(uri string, nonce string) (string, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("nonce", nonce)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}
	tr := &tokenResponse{}
	if err := json.Unmarshal(b, tr); err != nil {
		return "", err
	}

	return tr.Token, nil
}

func getAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("localhost", strconv.Itoa(0)))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := listener.Close(); err != nil {
			// ignore
			fmt.Printf("error closing listener: %v", err)
		}
	}()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
