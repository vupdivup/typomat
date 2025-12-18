package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// sessionKeyMap defines key bindings for the typing session UI.
type sessionKeyMap struct {
	Quit key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k sessionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k sessionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// sessionKeys holds the key bindings for the typing session UI.
var sessionKeys = sessionKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
}

// breakKeyMap defines key bindings for the break screen UI.
type breakKeyMap struct {
	Quit    key.Binding
	Restart key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k breakKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Restart}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k breakKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// breakKeys holds the key bindings for the break screen UI.
var breakKeys = breakKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
	Restart: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "retry"),
	),
}
