package listview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/config"
	"github.com/fcarp10/archutils/internal/scripts"
	helpkeys "github.com/fcarp10/archutils/internal/tui/helpkeys"
	"github.com/fcarp10/archutils/internal/tui/logsview"
)

const (
	stageMenu = iota
	stageCategory
	stageItems
	stageInstalling
)

const (
	menuPackages = iota
	menuInstallParu
	menuVSCodeExtensions
	menuAutologin
	menuPasswordlessSSH
	menuPasswordlessSudo
)

var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFF")).
			Padding(1)
	logsStyle             = listStyle.Align(lipgloss.Left, lipgloss.Center)
	listItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("51"))
)

type Model struct {
	width            int
	height           int
	logsView         logsview.Model
	categories       []config.Category
	categoryNames    []string
	selectedCategory config.Category
	cursor           int
	currentStage     int
	selectedItems    map[int]struct{}
	itemNames        []string
	logsVisible      bool
	directory        string
	installer        scripts.Installer
}

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

func (m Model) Init() tea.Cmd {
	return m.logsView.Init()
}

func New(installer scripts.Installer) Model {
	menuItemsTitles = make([]string, len(menuItems))
	for i, item := range menuItems {
		menuItemsTitles[i] = item.title
	}
	return Model{
		cursor:       0,
		currentStage: stageMenu,
		installer:    installer,
	}
}

func (m Model) SelectionCount() (selected, total int) {
	if m.currentStage != stageItems && m.currentStage != stageInstalling {
		return -1, -1
	}
	return len(m.selectedItems), len(m.itemNames)
}

func initCategories(dir string) ([]config.Category, []string, error) {
	categories, err := config.ReadCategories(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config directory %s: %w", dir, err)
	}
	return categories, config.CategoryNames(categories), nil
}

func initializeSelection(items []string) ([]string, map[int]struct{}) {
	selected := make(map[int]struct{})
	processed := make([]string, len(items))
	copy(processed, items)
	for i := range processed {
		if strings.HasPrefix(processed[i], "#") {
			processed[i] = strings.TrimSpace(strings.TrimPrefix(processed[i], "#"))
		} else {
			selected[i] = struct{}{}
		}
	}
	return processed, selected
}

func (m Model) showInformation() Model {
	m.logsVisible = true
	switch m.currentStage {
	case stageMenu:
		m.logsView = logsview.NewInfo(menuItems[m.cursor].description)
	case stageItems:
		description := m.selectedCategory.Items[m.cursor].Description
		if description == "" {
			description = "No information available for this item"
		}
		m.logsView = logsview.NewInfo(description)
	default:
		m.logsVisible = false
	}
	return m
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

func (m Model) handleCategoryEnter() (Model, tea.Cmd) {
	var names []string
	for _, item := range m.categories[m.cursor].Items {
		names = append(names, item.Name)
	}
	m.itemNames = names
	m.itemNames, m.selectedItems = initializeSelection(names)
	for i, item := range m.itemNames {
		switch m.directory {
		case config.PkgsDir():
			m.categories[m.cursor].Items[i].Description = m.installer.GetPackageDescription(item)
		case config.ExtDir():
			// TO-DO
		}
	}
	m.selectedCategory = m.categories[m.cursor]
	m.cursor = 0
	m.currentStage = stageItems
	m = m.showInformation()
	return m, nil
}

func (m Model) handleInstall() (Model, tea.Cmd) {
	if m.currentStage != stageItems {
		return m, nil
	}

	m.logsVisible = true
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

func (m Model) handleSelectAll() Model {
	if m.currentStage != stageItems {
		return m
	}
	m.selectedItems = make(map[int]struct{})
	for i := range m.itemNames {
		m.selectedItems[i] = struct{}{}
	}
	return m
}

func (m Model) handleDeselectAll() Model {
	if m.currentStage != stageItems {
		return m
	}
	m.selectedItems = make(map[int]struct{})
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, helpkeys.Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m = m.showInformation()
		case key.Matches(msg, helpkeys.Keys.Down):
			var listMenuLength int
			switch m.currentStage {
			case stageMenu:
				listMenuLength = len(menuItems)
			case stageCategory:
				listMenuLength = len(m.categoryNames)
			case stageItems:
				listMenuLength = len(m.itemNames)
			}
			if m.cursor < listMenuLength-1 {
				m.cursor++
			}
			m = m.showInformation()
		case key.Matches(msg, helpkeys.Keys.Enter):
			switch m.currentStage {
			case stageMenu:
				m, cmd = m.handleMenuEnter()
				return m, cmd
			case stageCategory:
				m, cmd = m.handleCategoryEnter()
				return m, cmd
			case stageItems:
				if _, ok := m.selectedItems[m.cursor]; ok {
					delete(m.selectedItems, m.cursor)
				} else {
					m.selectedItems[m.cursor] = struct{}{}
				}
			}
		case key.Matches(msg, helpkeys.Keys.Install):
			m, cmd = m.handleInstall()
			cmds = append(cmds, cmd)
		case key.Matches(msg, helpkeys.Keys.SelectAll):
			m = m.handleSelectAll()
		case key.Matches(msg, helpkeys.Keys.DeselectAll):
			m = m.handleDeselectAll()
		case key.Matches(msg, helpkeys.Keys.Back):
			if m.currentStage > 0 {
				m.itemNames = nil
				m.selectedCategory = config.Category{}
				m.selectedItems = make(map[int]struct{})
				m.currentStage = m.currentStage - 1
			}
			m.cursor = 0
			m = m.showInformation()
			return m, nil
		case key.Matches(msg, helpkeys.Keys.Quit):
			return m, tea.Quit
		}
	case logsview.DisableLogs:
		m.currentStage = stageItems
		m.logsVisible = false
	case tea.WindowSizeMsg:
		m.logsVisible = true
		m = m.showInformation()
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	default:
		m.logsView, cmd = m.logsView.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var currentList []string
	switch m.currentStage {
	case stageMenu:
		currentList = menuItemsTitles
	case stageCategory:
		currentList = m.categoryNames
	case stageItems:
		currentList = m.itemNames
	case stageInstalling:
		currentList = m.itemNames
	}
	var list string
	for i, choice := range currentList {
		cursor := " "
		choice = " " + choice
		if m.cursor == i {
			cursor = listItemSelectedStyle.Render("❯")
			choice = listItemSelectedStyle.Render(choice)
		}

		// Always render checkboxes in items and installing stages
		if m.currentStage == stageItems || m.currentStage == stageInstalling {
			checked := " "
			if _, ok := m.selectedItems[i]; ok {
				checked = "x"
			}

			checked = lipgloss.NewStyle().Render(" [" + checked + "]")
			list += fmt.Sprintf("%s%s%s\n", cursor, checked, choice)
		} else {
			list += fmt.Sprintf("%s%s\n", cursor, choice)
		}
	}
	list = strings.TrimRight(list, "\n")
	var s string
	if m.logsVisible {
		s = lipgloss.JoinHorizontal(lipgloss.Top, listStyle.Render(list), logsStyle.Render(m.logsView.View()))
	} else {
		s = listStyle.Render(list)
	}
	return s
}
