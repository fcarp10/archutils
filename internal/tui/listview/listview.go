package listview

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/packages"
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
	categories    []packages.CategoryPkgs
	categoryNames []string
	cursor        int
	listStage     int
	pkgsSelected  map[int]struct{}
	pkgs          []string
	extensions    bool
	logsVisible   bool
}

var menuItems = []string{
	"Install Packages",
	"Install Paru",
	"Pull Dotfiles",
}

func (m model) Init() tea.Cmd {
	return m.logsView.Init()
}

func New() model {
	categoriesPkgs, err := packages.ReadCategoriesPkgs()
	if err != nil {
		log.Fatalf("Failed to prepare package config files: %v", err)
	}
	return model{
		categories:    categoriesPkgs,
		categoryNames: packages.CategoryNames(categoriesPkgs),
		cursor:        0,
		listStage:     0,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, helpkeys.Keys.Up):
			m.logsVisible = false
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, helpkeys.Keys.Down):
			m.logsVisible = false
			var listMenuLength int
			switch m.listStage {
			case 0:
				listMenuLength = len(menuItems)
			case 1:
				listMenuLength = len(m.categoryNames)
			case 2:
				listMenuLength = len(m.pkgs)
			}
			if m.cursor < listMenuLength-1 {
				m.cursor++
			}
		case key.Matches(msg, helpkeys.Keys.Enter):
			switch m.listStage {
			case 0: // In main menu
				switch m.cursor {
				case 0:
					m.listStage = 1 // Move to package categories list
					m.cursor = 0
					return m, nil
				case 1: // Install Paru
					m.logsVisible = true
					m.logsView = logsview.NewScript()
					m.logsView, cmd = m.logsView.Update(logsview.RunningScript("paru"))
					cmds = append(cmds, cmd)
				case 2: // TO-DO
				}
			case 1: // In package categories list, select packages from a category
				m.extensions = false
				m.pkgsSelected = make(map[int]struct{})
				m.pkgs = make([]string, len(m.categories[m.cursor].PackagesNames))
				if strings.Contains(m.categories[m.cursor].Key, "extensions") {
					m.extensions = true
				}
				copy(m.pkgs, m.categories[m.cursor].PackagesNames)
				for i := 0; i < len(m.pkgs); i++ {
					if strings.HasPrefix(m.pkgs[i], "#") { // Select all packages without #
						m.pkgs[i] = strings.TrimSpace(strings.TrimPrefix(m.pkgs[i], "#"))
					} else {
						m.pkgsSelected[i] = struct{}{}
					}
				}
				m.listStage = 2
				m.cursor = 0
				return m, nil
			case 2: // In packages list, enable toggle selection
				if _, ok := m.pkgsSelected[m.cursor]; ok {
					delete(m.pkgsSelected, m.cursor)
				} else {
					m.pkgsSelected[m.cursor] = struct{}{}
				}
			}
		case key.Matches(msg, helpkeys.Keys.Install):
			m.logsVisible = true
			m.logsView = logsview.NewPackages(m.pkgsSelected, m.pkgs, m.extensions)
			if m.listStage == 2 { // If packages list, then install
				m.listStage = 3
				m.logsView, cmd = m.logsView.Update(msg)
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, helpkeys.Keys.Back):
			if m.listStage > 0 {
				// Move to previous list stage
				m.listStage = m.listStage - 1
				m.cursor = 0
				m.pkgsSelected = make(map[int]struct{})
				return m, nil
			}
		case key.Matches(msg, helpkeys.Keys.Quit):
			return m, tea.Quit
		}
	case logsview.DisableLogs:
		m.listStage = 2
		m.logsVisible = false
	case tea.WindowSizeMsg:
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
		currentList = menuItems
	case 1:
		currentList = m.categoryNames
	case 2:
		currentList = m.pkgs
	case 3:
		currentList = m.pkgs
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
			if len(m.pkgsSelected) > 0 {
				checked := " "
				if _, ok := m.pkgsSelected[i]; ok {
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
