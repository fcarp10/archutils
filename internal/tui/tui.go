package tui

import (
	"strings"

	hlp "github.com/fcarp10/archutils/internal/tui/helpkeys"
	"github.com/fcarp10/archutils/internal/tui/listview"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type mainModel struct {
	listView tea.Model
	help     help.Model
}

func InitialModel() mainModel {
	return mainModel{
		help:     help.New(),
		listView: listview.New(),
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
			m.listView, cmd = m.listView.Update(msg)
			cmds = append(cmds, cmd)
		}
	default:
		m.listView, cmd = m.listView.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	return "\n" + m.listView.View() + strings.Repeat("\n", 2) + m.help.View(hlp.Keys)
}
