package tui

import (
	"fmt"
	"strings"

	"github.com/fcarp10/archutils/internal/scripts"
	hlp "github.com/fcarp10/archutils/internal/tui/helpkeys"
	"github.com/fcarp10/archutils/internal/tui/listview"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var selectionCountStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))

type mainModel struct {
	listView listview.Model
	help     help.Model
}

func InitialModel() mainModel {
	return mainModel{
		help:     help.New(),
		listView: listview.New(scripts.Runner{}),
	}
}

func (m mainModel) Init() tea.Cmd {
	return m.listView.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, hlp.Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		default:
			var updated tea.Model
			updated, cmd = m.listView.Update(msg)
			if lv, ok := updated.(listview.Model); ok {
				m.listView = lv
			}
			cmds = append(cmds, cmd)
		}
	default:
		var updated tea.Model
		updated, cmd = m.listView.Update(msg)
		if lv, ok := updated.(listview.Model); ok {
			m.listView = lv
		}
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var statusBar string
	selected, total := m.listView.SelectionCount()
	if selected >= 0 && total >= 0 {
		statusBar = selectionCountStyle.Render(fmt.Sprintf("  Selected: %d/%d", selected, total))
	}

	helpView := m.help.View(hlp.Keys)
	content := "\n" + m.listView.View()
	if statusBar != "" {
		content += "\n" + statusBar
	}
	content += strings.Repeat("\n", 2) + helpView
	return content
}
