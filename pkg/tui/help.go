package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Tab  key.Binding
	// Left  key.Binding
	// Right key.Binding
	Help  key.Binding
	Quit  key.Binding
	Enter key.Binding
	Space key.Binding
	E     key.Binding
	C     key.Binding
	Num   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send request"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	E: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit file"),
	),
	C: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "edit config"),
	),
	Num: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("1-9", "select file"),
	),
}
