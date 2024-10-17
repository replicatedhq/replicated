package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

func (r *Replicated) Validate(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) error {
	if err := validateSecurity(ctx, source); err != nil {
		return err
	}

	if err := validateFunctionality(ctx, source); err != nil {
		return err
	}

	if err := validateCompatibility(ctx, source); err != nil {
		return err
	}

	if err := validatePerformance(ctx, source); err != nil {
		return err
	}

	return nil
}
