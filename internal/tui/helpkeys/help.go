package helpkeys

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Install key.Binding
	Help    key.Binding
	Quit    key.Binding
	Back    key.Binding
}

var Keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "Up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "Down"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "left"),
		key.WithHelp("←/esc", "Back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("⏎/␣", "Select"),
	),
	Install: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "Install packages"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "Toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "Exit app"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Back},
		{k.Enter, k.Install},
		{k.Help, k.Quit},
	}
}
