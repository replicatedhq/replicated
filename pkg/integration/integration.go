package integration

import (
	"path"
)

type format string

const (
	FormatJSON  format = "json"
	FormatTable format = "table"
)

func CLIPath() string {
	return path.Join("..", "..", "bin", "replicated")
}
