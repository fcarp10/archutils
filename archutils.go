package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	c "github.com/fcarp10/archutils/internal/config"
	"github.com/fcarp10/archutils/internal/tui"
)

// version is set at build time via -ldflags, e.g.:
//
//	go build -ldflags "-X main.version=v1.2.3"
var version = "dev"

//go:embed configs/*
var configFS embed.FS

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	showHelp := flag.Bool("help", false, "Print this help message")
	flag.BoolVar(showHelp, "h", false, "Print this help message (shorthand)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("archutils %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Printf(`archutils — Arch Linux Utilities TUI

Usage:
  archutils [flags]

Flags:
  --version   Print version and exit
  --help, -h  Print this help message

The TUI guides you through installing Arch Linux packages, VSCode
extensions, and system configurations interactively.

Environment variables:
  ARCHUTILS_EDITOR   Editor binary for extension management (default: codium)
`)
		os.Exit(0)
	}

	c.Init(configFS)

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
