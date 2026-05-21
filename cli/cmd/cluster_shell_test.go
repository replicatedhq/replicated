package cmd

import (
	"testing"
)

func TestGetShellSetupCommand(t *testing.T) {
	tests := []struct {
		name           string
		shellCmd       string
		kubeconfigPath string
		want           string
	}{
		{
			name:           "bash",
			shellCmd:       "/bin/bash",
			kubeconfigPath: "/tmp/replicated-kubeconfig123",
			want:           "export KUBECONFIG=/tmp/replicated-kubeconfig123\n",
		},
		{
			name:           "zsh",
			shellCmd:       "/usr/bin/zsh",
			kubeconfigPath: "/tmp/replicated-kubeconfig456",
			want:           "export KUBECONFIG=/tmp/replicated-kubeconfig456\n",
		},
		{
			name:           "fish",
			shellCmd:       "/usr/local/bin/fish",
			kubeconfigPath: "/tmp/replicated-kubeconfig789",
			want:           "set -x KUBECONFIG /tmp/replicated-kubeconfig789\n",
		},
		{
			name:           "nushell",
			shellCmd:       "/usr/bin/nu",
			kubeconfigPath: "/tmp/replicated-kubeconfigabc",
			want:           "$env.KUBECONFIG = \"/tmp/replicated-kubeconfigabc\"\n",
		},
		{
			name:           "nushell with path containing nu",
			shellCmd:       "/home/user/.cargo/bin/nu",
			kubeconfigPath: "/tmp/replicated-kubeconfigdef",
			want:           "$env.KUBECONFIG = \"/tmp/replicated-kubeconfigdef\"\n",
		},
		{
			name:           "sh",
			shellCmd:       "/bin/sh",
			kubeconfigPath: "/tmp/replicated-kubeconfigghi",
			want:           "export KUBECONFIG=/tmp/replicated-kubeconfigghi\n",
		},
		{
			name:           "just binary name bash",
			shellCmd:       "bash",
			kubeconfigPath: "/tmp/test",
			want:           "export KUBECONFIG=/tmp/test\n",
		},
		{
			name:           "just binary name nu",
			shellCmd:       "nu",
			kubeconfigPath: "/tmp/test",
			want:           "$env.KUBECONFIG = \"/tmp/test\"\n",
		},
		{
			name:           "shell ending in nu but not nushell (e.g., gnu)",
			shellCmd:       "/usr/bin/gnu",
			kubeconfigPath: "/tmp/test",
			want:           "export KUBECONFIG=/tmp/test\n",
		},
		{
			name:           "menu shell falls back to posix",
			shellCmd:       "/usr/bin/menu",
			kubeconfigPath: "/tmp/test",
			want:           "export KUBECONFIG=/tmp/test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShellSetupCommand(tt.shellCmd, tt.kubeconfigPath)
			if got != tt.want {
				t.Errorf("getShellSetupCommand(%q, %q) = %q, want %q", tt.shellCmd, tt.kubeconfigPath, got, tt.want)
			}
		})
	}
}
