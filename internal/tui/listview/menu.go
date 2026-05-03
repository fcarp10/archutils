package listview

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fcarp10/archutils/internal/config"
	"github.com/fcarp10/archutils/internal/tui/logsview"
)

type menuItem struct {
	title       string
	description string
}

var menuItems = []menuItem{
	{
		title:       "Arch Linux Packages",
		description: "A categorized collection of Arch Linux packages",
	},
	{
		title:       "Install Paru",
		description: "Paru AUR helper - a package manager for the Arch Linux community repository",
	},
	{
		title:       "VSCode Extensions",
		description: "A collection of VSCode extensions",
	},
	{
		title:       "Enable Autologin",
		description: "Enable and configure autologin for the current user",
	},
	{
		title:       "Enable Passwordless SSH",
		description: "Disable SSH password authentication and enable/restart SSH service",
	},
	{
		title:       "Configure Passwordless Sudo",
		description: "Configure passwordless sudo for the current user (will prompt for password once)",
	},
}

var menuItemsTitles []string

func (m Model) viewMenu() string {
	var list string
	total := len(menuItemsTitles)
	start, end := m.visibleRange(total)

	if start > 0 {
		list += scrollUpStyle.Render(fmt.Sprintf("  ▲ %d more", start)) + "\n"
	}
	for i := start; i < end; i++ {
		choice := menuItemsTitles[i]
		cursor := " "
		displayChoice := " " + choice
		if m.cursor == i {
			cursor = listItemSelectedStyle.Render("❯")
			displayChoice = listItemSelectedStyle.Render(displayChoice)
		}
		list += fmt.Sprintf("%s%s\n", cursor, displayChoice)
	}
	if end < total {
		list += scrollDownStyle.Render(fmt.Sprintf("  ▼ %d more", total-end)) + "\n"
	}
	return list
}

func (m Model) handleMenuEnter() (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.cursor {
	case menuInstallParu:
		m.logsVisible = true
		m.logsView = logsview.NewScript(m.installer)
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptParu))
		cmds = append(cmds, cmd)
	case menuAutologin:
		m.logsVisible = true
		m.logsView = logsview.NewScript(m.installer)
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptAutologin))
		cmds = append(cmds, cmd)
	case menuPasswordlessSSH:
		m.logsVisible = true
		m.logsView = logsview.NewScript(m.installer)
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptPasswordlessSSH))
		cmds = append(cmds, cmd)
	case menuPasswordlessSudo:
		m.logsVisible = true
		m.logsView = logsview.NewScript(m.installer)
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptPasswordlessSudo))
		cmds = append(cmds, cmd)
	default:
		switch m.cursor {
		case menuPackages:
			m.directory = config.PkgsDir()
		case menuVSCodeExtensions:
			m.directory = config.ExtDir()
		}
		var err error
		m.categories, m.categoryNames, err = initCategories(m.directory)
		if err != nil {
			m.logsVisible = true
			m.logsView = logsview.NewInfo(fmt.Sprintf("Error: %v", err))
			return m, nil
		}
		m.cursor = 0
		m.currentStage = stageCategory
		m = m.showInformation()
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
