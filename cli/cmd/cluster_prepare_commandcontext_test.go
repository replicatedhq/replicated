package cmd

import (
	"os"
	"strings"
	"testing"
)

// Structural check: cluster prepare must launch child processes with CommandContext
// so cobra cancellation (Ctrl+C) reaches kots install and the kots install script.
func TestClusterPrepareUsesCommandContext(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("cluster_prepare.go")
	if err != nil {
		t.Fatalf("read cluster_prepare.go: %v", err)
	}
	content := string(src)

	if !strings.Contains(content, `exec.CommandContext(ctx, kotCLI, "install"`) {
		t.Fatal("installKotsApp must use exec.CommandContext for kots install")
	}
	if !strings.Contains(content, `exec.CommandContext(ctx, installScript.Name())`) {
		t.Fatal("installKotsCLI must use exec.CommandContext for the kots install script")
	}
	if strings.Contains(content, `exec.Command(kotCLI, "install"`) {
		t.Fatal("installKotsApp must not use context-less exec.Command for kots install")
	}
	if strings.Contains(content, `exec.Command(installScript.Name())`) {
		t.Fatal("installKotsCLI must not use context-less exec.Command for the install script")
	}
}
