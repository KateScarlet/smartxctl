package main

import (
	"bytes"
	"strings"
	"testing"

	"smartxctl/internal/cli"
)

func TestRootCommandWritesErrors(t *testing.T) {
	cmd := cli.NewRootCmd()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"get", "vm"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected command error")
	}

	// main() prints the returned error to stderr, so Execute must surface it.
	if !strings.Contains(err.Error(), "specify a VM name or use --all") {
		t.Fatalf("unexpected command error: %v", err)
	}
}
