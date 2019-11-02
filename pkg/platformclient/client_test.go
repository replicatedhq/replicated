package platformclient

import (
	"os"
	"path"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

var (
	pact dsl.Pact
)

func TestMain(m *testing.M) {
	pact = createPact()

	pact.Setup(true)

	code := m.Run()

	pact.WritePact()
	pact.Teardown()

	os.Exit(code)
}

func createPact() dsl.Pact {
	dir, _ := os.Getwd()

	pactDir := path.Join(dir, "..", "..", "pacts")
	logDir := path.Join(dir, "..", "..", "logs")

	return dsl.Pact{
		Consumer: "replicated-cli",
		Provider: "vendor-api",
		LogDir:   logDir,
		PactDir:  pactDir,
		LogLevel: "debug",
	}
}
