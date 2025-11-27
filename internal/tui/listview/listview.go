package listview

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/config"
	helpkeys "github.com/fcarp10/archutils/internal/tui/helpkeys"
	"github.com/fcarp10/archutils/internal/tui/logsview"
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

type model struct {
	width         int
	height        int
	logsView      logsview.Model
	categories    []config.Category
	categoryNames []string
	cursor        int
	listStage     int
	itemsSelected map[int]struct{}
	itemsNames    []string
	logsVisible   bool
	directory     string
}

type menuItem struct {
	index       int
	title       string
	description string
}

var menuItems = []menuItem{
	{
		index:       1,
		title:       "Arch Linux Packages",
		description: "A categorized collection of Arch Linux packages"},
	{
		index:       2,
		title:       "Install Paru",
		description: "Paru AUR helper - a package manager for the Arch Linux community repository"},
	{
		index:       3,
		title:       "VSCode Extensions",
		description: "A collection of VSCode extensions",
	},
}

var menuItemsTitles []string

func (m model) Init() tea.Cmd {
	return m.logsView.Init()
}

func New() model {
	for _, item := range menuItems {
		menuItemsTitles = append(menuItemsTitles, item.title)
	}
	return model{
		cursor:    0,
		listStage: 0,
	}
}

func initCategories(dir string) ([]config.Category, []string) {
	categories, err := config.ReadCategories(dir)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	return categories, config.CategoryNames(categories)
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, helpkeys.Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			if m.listStage == 0 {
				m.logsVisible = true
				m.logsView = logsview.NewInstructions(menuItems[m.cursor].description)
			} else {
				m.logsVisible = false
			}
		case key.Matches(msg, helpkeys.Keys.Down):
			var listMenuLength int
			switch m.listStage {
			case 0:
				listMenuLength = len(menuItems)
			case 1:
				listMenuLength = len(m.categoryNames)
			case 2:
				listMenuLength = len(m.itemsNames)
			}
			if m.cursor < listMenuLength-1 {
				m.cursor++
			}
			if m.listStage == 0 {
				m.logsVisible = true
				m.logsView = logsview.NewInstructions(menuItems[m.cursor].description)
			} else {
				m.logsVisible = false
			}
		case key.Matches(msg, helpkeys.Keys.Enter):
			switch m.listStage {
			case 0: // Select menu item
				switch m.cursor {
				case 1: // Install Paru
					m.logsVisible = true
					m.logsView = logsview.NewScript()
					m.logsView, cmd = m.logsView.Update(logsview.RunningScript(logsview.ScriptParu))
					cmds = append(cmds, cmd)
				default: // Move to categories
					switch m.cursor {
					case 0:
						m.directory = config.PKGS_DIR
					case 2:
						m.directory = config.EXT_DIR
					}
					m.cursor = 0
					m.listStage = 1
					m.categories, m.categoryNames = initCategories(m.directory)
					return m, nil
				}
			case 1: // Select category
				m.itemsNames = m.categories[m.cursor].ItemsNames
				m.itemsNames, m.itemsSelected = initializeSelection(m.itemsNames)
				m.cursor = 0
				m.listStage = 2
				return m, nil
			case 2: // Toggle item
				if _, ok := m.itemsSelected[m.cursor]; ok {
					delete(m.itemsSelected, m.cursor)
				} else {
					m.itemsSelected[m.cursor] = struct{}{}
				}
			}
		case key.Matches(msg, helpkeys.Keys.Install):
			m.logsVisible = true
			if m.listStage == 2 { // If in toggle list, then install
				var itemsNames []string
				for idx := range m.itemsSelected {
					if idx < len(m.itemsNames) {
						itemsNames = append(itemsNames, m.itemsNames[idx])
					}
				}
				m.listStage = 3
				var installType logsview.ItemsInstallType
				switch m.directory {
				case config.PKGS_DIR:
					installType = logsview.InstallPackages
				case config.EXT_DIR:
					installType = logsview.InstallExtensions
				}
				m.logsView = logsview.NewItems(itemsNames)
				m.logsView, cmd = m.logsView.Update(logsview.InstallItems(installType))
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, helpkeys.Keys.Back):
			if m.listStage > 0 {
				m.listStage = m.listStage - 1 // Move to category list
			}
			m.cursor = 0
			m.itemsSelected = make(map[int]struct{})
			return m, nil
		case key.Matches(msg, helpkeys.Keys.Quit):
			return m, tea.Quit
		}
	case logsview.DisableLogs:
		m.listStage = 2
		m.logsVisible = false
	case tea.WindowSizeMsg:
		m.logsVisible = true
		m.logsView = logsview.NewInstructions(menuItems[m.cursor].description)
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	default:
		m.logsView, cmd = m.logsView.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var currentList []string
	switch m.listStage {
	case 0:
		currentList = menuItemsTitles
	case 1:
		currentList = m.categoryNames
	case 2:
		currentList = m.itemsNames
	case 3:
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

		if m.listStage == 2 || m.listStage == 3 {
			if len(m.itemsSelected) > 0 {
				checked := " "
				if _, ok := m.itemsSelected[i]; ok {
					checked = "x"
				}

				checked = lipgloss.NewStyle().Render(" [" + checked + "]")
				list += fmt.Sprintf("%s%s%s\n", cursor, checked, choice)
			}
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
