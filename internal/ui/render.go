package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	"github.com/vupdivup/typelines/internal/config"
	"github.com/vupdivup/typelines/pkg/textutils"
)

// renderTitleBar renders the title bar of the application.
func renderTitleBar(m model) string {
	left := mutedStyle.Render("╭" + "──")

	// * after title indicates break state
	var breakMarker string
	switch m.appState {
	case StateBreak:
		breakMarker = bodyStyle.Render("*")
	default:
		breakMarker = ""
	}

	// The title 'sits' on the upper border line
	title := accentStyle.Render(" "+config.AppName) +
		mutedStyle.Render("()") + breakMarker + " "

	restWidth := windowOuterWidth - lipgloss.Width(left) - lipgloss.Width(title)
	right := mutedStyle.Render(strings.Repeat("─", restWidth-1) + "╮")

	return left + title + right
}

// renderHelp renders the help view based on the current mode.
func renderHelp(m model) string {
	var keyMap help.KeyMap

	if m.appState == StateSession || m.appState == StateReady {
		m.help.Styles.ShortDesc = mutedStyle
		keyMap = sessionKeys
	} else {
		m.help.Styles.ShortDesc = bodyStyle
		keyMap = breakKeys
	}
	return m.help.View(keyMap)
}

// renderStats renders the typing statistics.
func renderStats(m model) string {
	var labelStyle lipgloss.Style

	if m.appState == StateSession || m.appState == StateReady {
		labelStyle = mutedStyle
	} else {
		labelStyle = bodyStyle
	}

	wpmStr := fmt.Sprintf("%d", int(m.wpm))
	accStr := fmt.Sprintf("%d%%", int(m.accuracy))

	return accentStyle.Render(wpmStr) +
		labelStyle.Render(" WPM") +
		bodyStyle.Render(" • ") +
		accentStyle.Render(accStr) +
		labelStyle.Render(" ACC")
}

// renderStatusBar renders the status bar with help and stats.
func renderStatusBar(m model) string {
	help := renderHelp(m)
	stats := renderStats(m)

	space := windowContentWidth - lipgloss.Width(help) - lipgloss.Width(stats)
	return help + lipgloss.NewStyle().Width(space).Render("") + stats
}

// renderPrompt renders the prompt with appropriate styles.
func renderPrompt(m model) string {
	render := ""

	promptLines := textutils.Wrap(m.prompt, canvasContentWidth, ' ')
	inputRunes := []rune(m.input)
	pos := 0

	for lineIdx, line := range promptLines {
		for _, promptChar := range line {
			var style lipgloss.Style

			switch m.appState {
			case StateBreak:
				// In break state, show all characters as muted except mistakes
				if len(m.input) <= pos {
					style = mutedStyle
				} else if promptChar != inputRunes[pos] {
					style = errorStyle
				} else if m.mistakes[pos] {
					style = accentStyle
				} else {
					style = mutedStyle
				}
			default:
				// In session state, style based on cursor and mistakes
				// Corrected mistakes are shown in accent color
				if pos > m.cursor() {
					if promptChar == ' ' {
						style = mutedStyle
					} else {
						style = bodyStyle
					}
				} else if pos == m.cursor() {
					style = bodyStyle.Underline(true)
				} else {
					if promptChar != inputRunes[pos] {
						style = errorStyle
					} else if m.mistakes[pos] {
						style = accentStyle
					} else {
						style = mutedStyle
					}
				}
			}

			if promptChar == ' ' {
				promptChar = '·'
			}

			render += style.Inline(true).Render(string(promptChar))
			pos++
		}

		if lineIdx < len(promptLines)-1 {
			render += "\n"
		}
	}

	return render
}

// renderLoad renders the loading indicator.
func renderLoad(m model) string {
	return m.spinner.View() + bodyStyle.Render(" Coming up with words...")
}

// renderCanvas renders the main canvas area based on the application state.
func renderCanvas(m model) string {
	switch m.appState {
	case StateLoading:
		return canvasStyle.Render(renderLoad(m))
	default:
		return canvasStyle.Render(renderPrompt(m))
	}
}

// renderWindow renders the main application window.
func renderWindow(m model) string {
	var statusBar string

	switch m.appState {
	case StateLoading:
		statusBar = ""
	default:
		statusBar = renderStatusBar(m)
	}

	return windowStyle.Render(
		renderCanvas(m) + "\n" + statusBar)
}

// renderApp renders the entire application UI.
func renderApp(m model) string {
	return "\n" + renderTitleBar(m) + "\n" + renderWindow(m) + "\n"
}
