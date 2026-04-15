package helpkeys

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Install       key.Binding
	SelectAll     key.Binding
	DeselectAll   key.Binding
	Search        key.Binding
	ConfirmYes    key.Binding
	ConfirmNo     key.Binding
	CancelInstall key.Binding
	Help          key.Binding
	Quit          key.Binding
	Back          key.Binding
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
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "Select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "Deselect all"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Search"),
	),
	ConfirmYes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "Confirm install"),
	),
	ConfirmNo: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "Cancel install"),
	),
	CancelInstall: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "Cancel install"),
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
		{k.Enter, k.SelectAll, k.DeselectAll},
		{k.Search, k.Install, k.CancelInstall},
		{k.ConfirmYes, k.ConfirmNo, k.Help},
		{k.Quit},
	}
}
