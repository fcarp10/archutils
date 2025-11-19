package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fcarp10/archutils/internal/packages"
	"github.com/fcarp10/archutils/internal/tui"
)

//go:embed configs/arch-pkgs/*.txt
var packagesFS embed.FS

func main() {

	packages.PackagesFS = packagesFS

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
