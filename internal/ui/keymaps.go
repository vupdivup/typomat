package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// globalKeyMap defines global key bindings for the TUI.
type globalKeyMap struct {
	Quit key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// globalKeys holds the global key bindings for the TUI.
var globalKeys = globalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("esc", "quit"),
	),
}

// sessionKeyMap defines key bindings for the typing session UI.
type sessionKeyMap struct {
	Quit key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k sessionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{globalKeys.Quit}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k sessionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// sessionKeys holds the key bindings for the typing session UI.
var sessionKeys = sessionKeyMap{}

// breakKeyMap defines key bindings for the break screen UI.
type breakKeyMap struct {
	Restart key.Binding
}

// ShortHelp returns key bindings to be shown in the mini help view.
func (k breakKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{globalKeys.Quit, k.Restart}
}

// FullHelp returns key bindings to be shown in the expanded help view.
func (k breakKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

// breakKeys holds the key bindings for the break screen UI.
var breakKeys = breakKeyMap{
	Restart: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "retry"),
	),
}
