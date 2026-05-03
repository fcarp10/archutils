package listview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/tui/logsview"
)

func (m Model) viewItems() string {
	var list string
	indices := m.getFilteredIndices()
	if len(indices) == 0 && m.searchQuery != "" {
		list = noMatchStyle.Render("  No matching items")
	} else {
		for displayIdx, origIdx := range indices {
			choice := m.itemNames[origIdx]
			cursor := " "
			displayChoice := " " + choice

			if m.installedItems[origIdx] {
				displayChoice = installedItemStyle.Render(displayChoice + " ✓")
			}

			if m.cursor == displayIdx {
				cursor = listItemSelectedStyle.Render("❯")
				displayChoice = listItemSelectedStyle.Render(displayChoice)
			}

			checked := " "
			if _, ok := m.selectedItems[origIdx]; ok {
				checked = "x"
			}
			checked = lipgloss.NewStyle().Render(" [" + checked + "]")
			list += fmt.Sprintf("%s%s%s\n", cursor, checked, displayChoice)
		}
	}
	return strings.TrimRight(list, "\n")
}

func (m Model) handleInstall() (Model, tea.Cmd) {
	if m.currentStage != stageItems {
		return m, nil
	}

	m.searchMode = false
	m.searchQuery = ""

	if len(m.selectedItems) == 0 {
		m.logsVisible = true
		m.logsView = logsview.NewInfo("No items selected. Select items with space/enter.")
		return m, nil
	}

	m.currentStage = stageConfirm
	m.logsVisible = true

	var selectedList []string
	for idx := range m.selectedItems {
		if idx < len(m.itemNames) {
			selectedList = append(selectedList, m.itemNames[idx])
		}
	}

	confirmMsg := fmt.Sprintf("Confirm installation of %d item(s):\n\n", len(selectedList))
	for _, name := range selectedList {
		confirmMsg += "  • " + name + "\n"
	}
	confirmMsg += "\n  y: Confirm   n: Cancel"

	m.logsView = logsview.NewInfo(confirmMsg)
	return m, nil
}

func (m Model) handleSelectAll() Model {
	if m.currentStage != stageItems {
		return m
	}
	indices := m.getFilteredIndices()
	for _, idx := range indices {
		m.selectedItems[idx] = struct{}{}
	}
	return m
}

func (m Model) handleDeselectAll() Model {
	if m.currentStage != stageItems {
		return m
	}
	indices := m.getFilteredIndices()
	for _, idx := range indices {
		delete(m.selectedItems, idx)
	}
	return m
}
