package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	c "github.com/fcarp10/archutils/internal/config"
	"github.com/fcarp10/archutils/internal/tui"
)

//go:embed configs/*
var configFS embed.FS

func main() {

	c.ConfigFS = configFS

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
