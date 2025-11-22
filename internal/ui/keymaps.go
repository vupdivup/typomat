package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// sessionKeyMap defines key bindings for the typing session UI.
type sessionKeyMap struct {
	Stop key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k sessionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Stop}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k sessionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Stop},
	}
}

// sessionKeys holds the key bindings for the typing session UI.
var sessionKeys = sessionKeyMap{
	Stop: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "stop"),
	),
}

// breakKeyMap defines key bindings for the break screen UI.
type breakKeyMap struct {
	Restart key.Binding
	Quit    key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k breakKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Restart, k.Quit}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k breakKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Restart},
		{k.Quit},
	}
}

// breakKeys holds the key bindings for the break screen UI.
var breakKeys = breakKeyMap{
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
}
