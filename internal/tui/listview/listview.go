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

// Navigation stages
const (
	stageMenu = iota
	stageCategory
	stageItems
	stageInstalling
)

// Menu item indices
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

// Model is the main list view model exported for use by the TUI package.
type Model struct {
	width            int
	height           int
	logsView         logsview.Model
	categories       []config.Category
	categoryNames    []string
	categorySelected config.Category
	cursor           int
	listStage        int
	itemsSelected    map[int]struct{}
	itemsNames       []string
	logsVisible      bool
	directory        string
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

func New() Model {
	menuItemsTitles = make([]string, len(menuItems))
	for i, item := range menuItems {
		menuItemsTitles[i] = item.title
	}
	return Model{
		cursor:    0,
		listStage: stageMenu,
	}
}

// SelectionCount returns the number of selected items and total items.
// Returns -1 for total if not in the items stage.
func (m Model) SelectionCount() (selected, total int) {
	if m.listStage != stageItems && m.listStage != stageInstalling {
		return -1, -1
	}
	return len(m.itemsSelected), len(m.itemsNames)
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
	switch m.listStage {
	case stageMenu:
		m.logsView = logsview.NewInfo(menuItems[m.cursor].description)
	case stageItems:
		description := m.categorySelected.Items[m.cursor].Description
		if description == "" {
			description = "No information available for this item"
		}
		m.logsView = logsview.NewInfo(description)
	default:
		m.logsVisible = false
	}
	return m
}

// handleMenuEnter handles Enter key presses at the main menu stage.
func (m Model) handleMenuEnter() (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.cursor {
	case menuInstallParu:
		m.logsVisible = true
		m.logsView = logsview.NewScript()
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptParu))
		cmds = append(cmds, cmd)
	case menuAutologin:
		m.logsVisible = true
		m.logsView = logsview.NewScript()
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptAutologin))
		cmds = append(cmds, cmd)
	case menuPasswordlessSSH:
		m.logsVisible = true
		m.logsView = logsview.NewScript()
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptPasswordlessSSH))
		cmds = append(cmds, cmd)
	case menuPasswordlessSudo:
		m.logsVisible = true
		m.logsView = logsview.NewScript()
		m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptPasswordlessSudo))
		cmds = append(cmds, cmd)
	default:
		switch m.cursor {
		case menuPackages:
			m.directory = config.PKGS_DIR
		case menuVSCodeExtensions:
			m.directory = config.EXT_DIR
		}
		var err error
		m.categories, m.categoryNames, err = initCategories(m.directory)
		if err != nil {
			m.logsVisible = true
			m.logsView = logsview.NewInfo(fmt.Sprintf("Error: %v", err))
			return m, nil
		}
		m.cursor = 0
		m.listStage = stageCategory
		m = m.showInformation()
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

// handleCategoryEnter handles Enter key presses at the category selection stage.
func (m Model) handleCategoryEnter() (Model, tea.Cmd) {
	var names []string
	for _, item := range m.categories[m.cursor].Items {
		names = append(names, item.Name)
	}
	m.itemsNames = names
	m.itemsNames, m.itemsSelected = initializeSelection(names)
	for i, item := range m.itemsNames {
		switch m.directory {
		case config.PKGS_DIR:
			m.categories[m.cursor].Items[i].Description = scripts.GetPackageDescription(item)
		case config.EXT_DIR:
			// TO-DO
		}
	}
	m.categorySelected = m.categories[m.cursor]
	m.cursor = 0
	m.listStage = stageItems
	m = m.showInformation()
	return m, nil
}

// handleInstall handles the Install key press at the items stage.
func (m Model) handleInstall() (Model, tea.Cmd) {
	if m.listStage != stageItems {
		return m, nil
	}

	m.logsVisible = true
	var itemsNames []string
	for idx := range m.itemsSelected {
		if idx < len(m.itemsNames) {
			itemsNames = append(itemsNames, m.itemsNames[idx])
		}
	}
	m.listStage = stageInstalling
	var installType logsview.ItemsInstallType
	switch m.directory {
	case config.PKGS_DIR:
		installType = logsview.InstallPackages
	case config.EXT_DIR:
		installType = logsview.InstallExtensions
	}
	m.logsView = logsview.NewItems(itemsNames)
	var cmd tea.Cmd
	m.logsView, cmd = m.logsView.Update(logsview.InstallItems(installType))
	return m, cmd
}

// handleSelectAll selects all items in the current items list.
func (m Model) handleSelectAll() Model {
	if m.listStage != stageItems {
		return m
	}
	m.itemsSelected = make(map[int]struct{})
	for i := range m.itemsNames {
		m.itemsSelected[i] = struct{}{}
	}
	return m
}

// handleDeselectAll deselects all items in the current items list.
func (m Model) handleDeselectAll() Model {
	if m.listStage != stageItems {
		return m
	}
	m.itemsSelected = make(map[int]struct{})
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
			switch m.listStage {
			case stageMenu:
				listMenuLength = len(menuItems)
			case stageCategory:
				listMenuLength = len(m.categoryNames)
			case stageItems:
				listMenuLength = len(m.itemsNames)
			}
			if m.cursor < listMenuLength-1 {
				m.cursor++
			}
			m = m.showInformation()
		case key.Matches(msg, helpkeys.Keys.Enter):
			switch m.listStage {
			case stageMenu:
				m, cmd = m.handleMenuEnter()
				return m, cmd
			case stageCategory:
				m, cmd = m.handleCategoryEnter()
				return m, cmd
			case stageItems: // Toggle item
				if _, ok := m.itemsSelected[m.cursor]; ok {
					delete(m.itemsSelected, m.cursor)
				} else {
					m.itemsSelected[m.cursor] = struct{}{}
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
			if m.listStage > 0 {
				m.itemsNames = nil
				m.categorySelected = config.Category{}
				m.itemsSelected = make(map[int]struct{})
				m.listStage = m.listStage - 1 // Move to category list
			}
			m.cursor = 0
			m = m.showInformation()
			return m, nil
		case key.Matches(msg, helpkeys.Keys.Quit):
			return m, tea.Quit
		}
	case logsview.DisableLogs:
		m.listStage = stageItems
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
	switch m.listStage {
	case stageMenu:
		currentList = menuItemsTitles
	case stageCategory:
		currentList = m.categoryNames
	case stageItems:
		currentList = m.itemsNames
	case stageInstalling:
		currentList = m.itemsNames
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
		if m.listStage == stageItems || m.listStage == stageInstalling {
			checked := " "
			if _, ok := m.itemsSelected[i]; ok {
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
