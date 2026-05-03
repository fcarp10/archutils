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

// Stage identifiers for the UI state machine.
const (
	stageMenu = iota
	stageCategory
	stageItems
	stageConfirm
	stageInstalling
)

// Menu option indices.
const (
	menuPackages = iota
	menuInstallParu
	menuVSCodeExtensions
	menuAutologin
	menuPasswordlessSSH
	menuPasswordlessSudo
)

// Styling shared across all stages.
var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFF")).
			Padding(1)
	logsStyle             = listStyle.Align(lipgloss.Left, lipgloss.Center)
	listItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("51"))
	installedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
	searchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)
	noMatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// Model is the main list view model that manages all UI stages.
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
	installedItems   map[int]bool
	logsVisible      bool
	directory        string
	installer        scripts.Installer
	searchMode       bool
	searchQuery      string
}

// New creates a new Model starting at the main menu.
func New(installer scripts.Installer) Model {
	menuItemsTitles = make([]string, len(menuItems))
	for i, item := range menuItems {
		menuItemsTitles[i] = item.title
	}
	return Model{
		cursor:         0,
		currentStage:   stageMenu,
		installer:      installer,
		installedItems: make(map[int]bool),
	}
}

func (m Model) Init() tea.Cmd {
	return m.logsView.Init()
}

// SelectionCount returns the number of selected items and total items,
// or (-1, -1) if the current stage does not show selections.
func (m Model) SelectionCount() (selected, total int) {
	if m.currentStage != stageItems && m.currentStage != stageConfirm && m.currentStage != stageInstalling {
		return -1, -1
	}
	return len(m.selectedItems), len(m.itemNames)
}

func (m Model) getFilteredIndices() []int {
	if m.searchQuery == "" {
		indices := make([]int, len(m.itemNames))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}
	query := strings.ToLower(m.searchQuery)
	var indices []int
	for i, name := range m.itemNames {
		if strings.Contains(strings.ToLower(name), query) {
			indices = append(indices, i)
		}
	}
	return indices
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
	if m.logsView.IsActive() {
		return m
	}
	m.logsVisible = true
	switch m.currentStage {
	case stageMenu:
		m.logsView = logsview.NewInfo(menuItems[m.cursor].description)
	case stageItems:
		indices := m.getFilteredIndices()
		if len(indices) == 0 {
			m.logsView = logsview.NewInfo("No matching items")
		} else if m.cursor < len(indices) {
			origIdx := indices[m.cursor]
			description := m.selectedCategory.Items[origIdx].Description
			if description == "" {
				description = "No information available for this item"
			}
			m.logsView = logsview.NewInfo(description)
		}
	case stageConfirm:
	default:
		m.logsVisible = false
	}
	return m
}

// Update dispatches messages to the appropriate stage handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.currentStage == stageConfirm {
			switch {
			case key.Matches(msg, helpkeys.Keys.ConfirmYes):
				m, cmd = m.handleConfirmYes()
				cmds = append(cmds, cmd)
			case key.Matches(msg, helpkeys.Keys.ConfirmNo), key.Matches(msg, helpkeys.Keys.Back):
				m = m.handleConfirmNo()
			}
			return m, tea.Batch(cmds...)
		}

		if m.searchMode {
			return m.handleSearchInput(msg), nil
		}

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
				listMenuLength = len(m.getFilteredIndices())
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
				indices := m.getFilteredIndices()
				if m.cursor < len(indices) {
					origIdx := indices[m.cursor]
					if _, ok := m.selectedItems[origIdx]; ok {
						delete(m.selectedItems, origIdx)
					} else {
						m.selectedItems[origIdx] = struct{}{}
					}
				}
			}
		case key.Matches(msg, helpkeys.Keys.Search):
			if m.currentStage == stageItems {
				m.searchMode = true
				m.searchQuery = ""
				m.cursor = 0
			}
		case key.Matches(msg, helpkeys.Keys.Install):
			m, cmd = m.handleInstall()
			cmds = append(cmds, cmd)
		case key.Matches(msg, helpkeys.Keys.CancelInstall):
			if m.currentStage == stageInstalling {
				m.logsView, cmd = m.logsView.Update(logsview.CancelInstall{})
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, helpkeys.Keys.SelectAll):
			m = m.handleSelectAll()
		case key.Matches(msg, helpkeys.Keys.DeselectAll):
			m = m.handleDeselectAll()
		case key.Matches(msg, helpkeys.Keys.Back):
			if m.currentStage > 0 {
				m.itemNames = nil
				m.selectedCategory = config.Category{}
				m.selectedItems = make(map[int]struct{})
				m.installedItems = make(map[int]bool)
				m.searchMode = false
				m.searchQuery = ""
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
		m.searchMode = false
		m.searchQuery = ""
		if string(msg) != "" {
			m.logsVisible = true
			m.logsView = logsview.NewInfo(string(msg))
		} else {
			m.logsVisible = false
		}
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

func (m Model) handleSearchInput(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyRunes:
		if !msg.Alt {
			m.searchQuery += string(msg.Runes)
			m.cursor = 0
			m = m.showInformation()
		}
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.cursor = 0
			m = m.showInformation()
		}
	case tea.KeyEnter:
		m.searchMode = false
	case tea.KeyEscape:
		m.searchMode = false
		m.searchQuery = ""
		m.cursor = 0
		m = m.showInformation()
	default:
		switch {
		case key.Matches(msg, helpkeys.Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m = m.showInformation()
		case key.Matches(msg, helpkeys.Keys.Down):
			indices := m.getFilteredIndices()
			if m.cursor < len(indices)-1 {
				m.cursor++
			}
			m = m.showInformation()
		}
	}
	return m
}

// View renders the current stage.
func (m Model) View() string {
	var list string

	switch m.currentStage {
	case stageMenu:
		list = m.viewMenu()
	case stageCategory:
		list = m.viewCategory()
	case stageItems:
		list = m.viewItems()
	case stageConfirm, stageInstalling:
		list = m.viewConfirmInstalling()
	}

	if m.searchMode {
		prompt := searchPromptStyle.Render("/" + m.searchQuery + "▎")
		list = prompt + "\n" + list
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
