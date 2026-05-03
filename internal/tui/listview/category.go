package listview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fcarp10/archutils/internal/config"
)

func (m Model) viewCategory() string {
	var list string
	for i, choice := range m.categoryNames {
		cursor := " "
		displayChoice := " " + choice
		if m.cursor == i {
			cursor = listItemSelectedStyle.Render("❯")
			displayChoice = listItemSelectedStyle.Render(displayChoice)
		}
		list += fmt.Sprintf("%s%s\n", cursor, displayChoice)
	}
	return list
}

func (m Model) handleCategoryEnter() (Model, tea.Cmd) {
	var names []string
	for _, item := range m.categories[m.cursor].Items {
		names = append(names, item.Name)
	}
	m.itemNames = names
	m.itemNames, m.selectedItems = initializeSelection(names)
	m.installedItems = make(map[int]bool)

	switch m.directory {
	case config.PkgsDir():
		installed := m.installer.GetInstalledPackages()
		for i, item := range m.itemNames {
			fields := strings.Fields(item)
			if len(fields) == 0 {
				continue
			}
			pkgName := fields[0]
			if desc, ok := installed[pkgName]; ok {
				m.installedItems[i] = true
				m.categories[m.cursor].Items[i].Description = desc
			}
		}
	case config.ExtDir():
		for i, item := range m.itemNames {
			if m.installer.IsExtensionInstalled(item) {
				m.installedItems[i] = true
				m.categories[m.cursor].Items[i].Description = m.installer.GetExtensionDescription(item)
			}
		}
	}

	m.selectedCategory = m.categories[m.cursor]
	m.cursor = 0
	m.currentStage = stageItems
	m.searchMode = false
	m.searchQuery = ""
	m = m.showInformation()
	return m, nil
}
