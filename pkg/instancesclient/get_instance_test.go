package instancesclient

import (
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

var (
	pact dsl.Pact
)

func TestMain(m *testing.M) {
	pact = dsl.Pact{
		Consumer: "webapp",
		Provider: "api",
	}

	pact.Setup(true)

	rc := m.Run()

	pact.WritePact()
	pact.Teardown()

	os.Exit(rc)
}

func TestPactGetInstance(t *testing.T) {
	pact.
		AddInteraction().
		UponReceiving("greeting").
		WithRequest(dsl.Request{
			Method: http.MethodGet,
			Path:   dsl.String("/hello"),
		}).
		WillRespondWith(dsl.Response{
			Status: http.StatusOK,
			Headers: dsl.MapMatcher{
				"Content-Type": dsl.String("text/plain"),
			},
			Body: dsl.String("hello"),
		})

	if err := pact.Verify(func() error {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", pact.Host, pact.Server.Port, "hello"))
		if err != nil {
			return err
		}

		b, _ := io.ReadAll(resp.Body)
		if string(b) != "hello" {
			return fmt.Errorf("got %s expected %s", string(b), "hello")
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
