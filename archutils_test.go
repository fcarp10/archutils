package main

import (
	"flag"
	"os"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	// Save original os.Args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Reset flags between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"archutils", "--version"}
	showVersion := flag.Bool("version", false, "")
	flag.Parse()

	if !*showVersion {
		t.Error("expected --version flag to be true")
	}

	if version == "" {
		t.Error("expected version to be non-empty")
	}
}

func TestHelpFlag(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"archutils", "--help"}
	showHelp := flag.Bool("help", false, "")
	flag.Parse()

	if !*showHelp {
		t.Error("expected --help flag to be true")
	}
}

func TestHelpShorthand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"archutils", "-h"}
	showHelp := flag.Bool("help", false, "")
	flag.BoolVar(showHelp, "h", false, "")
	flag.Parse()

	if !*showHelp {
		t.Error("expected -h flag to be true")
	}
}

func TestVersionDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("expected default version 'dev', got %q", version)
	}
}
