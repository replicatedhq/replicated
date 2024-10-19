package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

func validatePerformance(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) (bool, map[string]Logs, error) {
	return true, map[string]Logs{}, nil
}
