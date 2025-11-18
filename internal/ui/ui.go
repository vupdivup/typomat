// ui provides a scrollable typing interface for text prompts.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Configuration
	scrollAnchorLeft = 20
	width = 80

	// Styles
	blockStyle = lipgloss.NewStyle().Width(width).Align(lipgloss.Left)
	baseStyle = lipgloss.NewStyle().Inline(true)
	normalStyle = baseStyle.Foreground(lipgloss.Color("7"))
	cursorStyle = normalStyle.Underline(true)
	correctStyle = baseStyle.Foreground(lipgloss.Color("2"))
	incorrectStyle = baseStyle.Foreground(lipgloss.Color("1"))
)

// model defines the TUI state.
type model struct {
	prompt   string
	input    string
}

// Cursor returns the current cursor position in the input string.
func (m model) Cursor() int {
	return len(m.input)
}

// Init is the initial command for the TUI.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the TUI state.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
		}
	}

	return m, nil
}

// View renders the TUI interface.
func (m model) View() string {
	var render string

	inputRunes := []rune(m.input)
	promptRunes := []rune(m.prompt)

	// Get rune indices for rendering based on cursor position
	renderStart := max(0, m.Cursor()-scrollAnchorLeft)
	renderEnd := min(len(promptRunes), renderStart+width)

	for i, r := range promptRunes {
		if i < renderStart || i >= renderEnd {
			continue
		}

		var style lipgloss.Style

		// Determine style based on cursor position and correctness
		if i == m.Cursor() {
			style = cursorStyle
		} else if i > m.Cursor() {
			style = normalStyle
		} else if r != inputRunes[i] {
			style = incorrectStyle
		} else {
			style = correctStyle
		}

		// Render spaces as middle dots for visibility
		if r == ' ' {
			render += style.Render("·")
		} else {
			render += style.Render(string(r))
		}
	}

	return blockStyle.Render(render)
}

// Launch starts the TUI with the given text prompt as a scrollable typing
// interface.
func Launch(prompt string) {
	p := tea.NewProgram(model{prompt: prompt})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
