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
	errChan := make(chan error)
	tokenChan := make(chan string)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := r.URL.Query().Get("nonce")
		if nonce == "" {
			errChan <- fmt.Errorf("no nonce found in response")
		}

		exchange := r.URL.Query().Get("exchange")
		if exchange == "" {
			errChan <- fmt.Errorf("no exchange found in response")
		}

		token, err := exchangeNonceForToken(exchange, nonce)
		if err != nil {
			errChan <- err
		}

		tokenChan <- token

		w.Write([]byte("Authentication successful. You may close this window."))
	})

	go func() {
		sm := http.NewServeMux()
		sm.Handle("/callback", handler)
		server := &http.Server{
			Addr:              net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
			Handler:           sm,
			ReadHeaderTimeout: 3 * time.Second,
		}

		errChan <- server.ListenAndServe()
	}()

	timeout := time.NewTicker(5 * time.Minute)

	select {
	case <-timeout.C:
		ctx.Done()
		return "", fmt.Errorf("authentication timeout")
	case token := <-tokenChan:
		return token, nil
	case e := <-errChan:
		ctx.Done()
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
