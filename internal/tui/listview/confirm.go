package listview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/config"
	"github.com/fcarp10/archutils/internal/tui/logsview"
)

func (m Model) viewConfirmInstalling() string {
	var list string
	total := len(m.itemNames)
	start, end := m.visibleRange(total)

	if start > 0 {
		list += scrollUpStyle.Render(fmt.Sprintf("  ▲ %d more", start)) + "\n"
	}
	for i := start; i < end; i++ {
		choice := m.itemNames[i]
		cursor := " "
		displayChoice := " " + choice

		if m.installedItems[i] {
			displayChoice = installedItemStyle.Render(displayChoice + " ✓")
		}

		if m.cursor == i {
			cursor = listItemSelectedStyle.Render("❯")
			displayChoice = listItemSelectedStyle.Render(displayChoice)
		}

		checked := " "
		if _, ok := m.selectedItems[i]; ok {
			checked = "x"
		}
		checked = lipgloss.NewStyle().Render(" [" + checked + "]")
		list += fmt.Sprintf("%s%s%s\n", cursor, checked, displayChoice)
	}
	if end < total {
		list += scrollDownStyle.Render(fmt.Sprintf("  ▼ %d more", total-end)) + "\n"
	}
	return strings.TrimRight(list, "\n")
}

func (m Model) handleConfirmYes() (Model, tea.Cmd) {
	m.logsVisible = true
	m.searchMode = false
	m.searchQuery = ""
	var selectedItemNames []string
	for idx := range m.selectedItems {
		if idx < len(m.itemNames) {
			selectedItemNames = append(selectedItemNames, m.itemNames[idx])
		}
	}
	m.currentStage = stageInstalling
	var installType logsview.ItemsInstallType
	switch m.directory {
	case config.PkgsDir():
		installType = logsview.InstallPackages
	case config.ExtDir():
		installType = logsview.InstallExtensions
	}
	m.logsView = logsview.NewItems(selectedItemNames, m.installer)
	var cmd tea.Cmd
	m.logsView, cmd = m.logsView.Update(logsview.InstallItems(installType))
	return m, cmd
}

func (m Model) handleConfirmNo() Model {
	m.currentStage = stageItems
	m.searchMode = false
	m.searchQuery = ""
	m.logsVisible = false
	return m
}
