// ui provides a text-based user interface (TUI) for the application.
package ui

import (
	"math"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vupdivup/recital/internal/domain"
	"github.com/vupdivup/recital/pkg/alphabet"
	"github.com/vupdivup/recital/pkg/metrics"
)

const (
	// Configuration
	width              = 80
	contentWidth       = width - 2
	windowPaddingV     = 1
	windowContentWidth = contentWidth - 2*windowPaddingV
	promptWidth        = windowContentWidth - 2*3
)

var (
	// Styles
	windowStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Padding(0, windowPaddingV).
			Border(lipgloss.RoundedBorder()).
			BorderTop(false).
			BorderForeground(mutedColor)

	promptStyle = lipgloss.NewStyle().
			Padding(1, 3).Height(4)

	accentColor = lipgloss.Color("75")
	bodyColor   = lipgloss.Color("252")
	mutedColor  = lipgloss.Color("244")
	errorColor  = lipgloss.Color("167")

	accentStyle = lipgloss.NewStyle().Foreground(accentColor)
	bodyStyle   = lipgloss.NewStyle().Foreground(bodyColor)
	mutedStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	errorStyle  = lipgloss.NewStyle().Foreground(errorColor)

	allowedInputRunes = append(alphabet.AllRunes, ' ')
)

type AppState int

const (
	StateLoading AppState = iota
	StateSession
	StateBreak
	StateReady
)

// model defines the TUI state.
type model struct {
	// dirPath is the directory path for prompts.
	dirPath string

	// state is the current application state.
	state AppState

	// prompt is the text prompt to type.
	prompt string
	// input is the current user input.
	input string
	// cursor is the current cursor position.
	cursor int
	// mistakes records the positions of mistakes made.
	mistakes map[int]bool

	// startTime is the time when the typing session started.
	startTime time.Time
	// wpm is the current words per minute.
	wpm int
	// accuracy is the current typing accuracy.
	accuracy int
	
	// help is the help view model.
	help help.Model
	// spinner is the loading spinner view model.
	spinner spinner.Model
}

type promptMsg string

// promptCmd fetches a new prompt from the domain layer.
func promptCmd(m model) tea.Cmd {
	return func() tea.Msg {
		prompt, err := domain.Prompt(m.dirPath, 128)
		if err != nil {
			panic(err) // TODO: graceful error handling
		}
		return promptMsg(prompt)
	}
}

// initialModel creates the initial TUI model.
func initialModel(dirPath string) model {
	help := help.New()
	help.Styles.ShortKey = accentStyle
	help.Styles.ShortSeparator = bodyStyle

	spinner := spinner.New(
		spinner.WithSpinner(spinner.Dot), spinner.WithStyle(accentStyle))

	m := model{
		dirPath: dirPath,
		help:    help,
		state:   StateLoading,
		spinner: spinner,
	}

	return m
}

// ready sets up the model for a ready state with a new prompt.
func (m *model) ready(prompt string) {
	m.prompt = prompt
	m.mistakes = make(map[int]bool)
	m.input = ""
	m.cursor = 0
	m.wpm = 0
	m.accuracy = 0.0
	m.state = StateReady
}

// start begins the typing session.
func (m *model) start() {
	m.state = StateSession
	m.startTime = time.Now()
}

// stop ends the typing session.
func (m *model) stop() {
	m.state = StateBreak
}

// Init is the initial command for the TUI.
func (m model) Init() tea.Cmd {
	return tea.Batch(promptCmd(m), m.spinner.Tick)
}

// Update handles incoming messages and updates the TUI state.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Quit if Ctrl+C is pressed
		if key.Matches(msg, globalKeys.Quit) {
			return m, tea.Quit
		}

		msgStr := msg.String()

		if m.state == StateSession || m.state == StateReady {
			if key.Matches(msg, sessionKeys.Stop) {
				m.stop()
				return m, nil
			}

			if msgStr == "backspace" && m.cursor > 0 {
				// Handle backspace
				m.cursor--
				m.input = m.input[:len(m.input)-1]
			} else {
				msgRunes := []rune(msgStr)
				promptRunes := []rune(m.prompt)

				// Ignore non-character keys or unsupported runes
				if len(msgRunes) != 1 ||
					!slices.Contains(allowedInputRunes, msgRunes[0]) {
					return m, nil
				}

				// Start session on first valid input
				if m.state == StateReady {
					m.start()
				}

				// Check for mistake
				if msgStr != string(promptRunes[m.cursor]) {
					m.mistakes[m.cursor] = true
				}

				// Accept input
				m.input += msgStr
				m.cursor++

				// Update metrics
				elapsed := time.Since(m.startTime)
				m.wpm = int(math.Round(metrics.WPM(m.input, elapsed)))
				m.accuracy = int(
					math.Round(metrics.Accuracy(m.prompt, m.input)))

				// End session if prompt completed
				if m.cursor >= len(promptRunes) {
					m.stop()
				}
			}
		} else {
			switch {
			case key.Matches(msg, breakKeys.Restart):
				m.state = StateLoading
				return m, tea.Batch(promptCmd(m), m.spinner.Tick)
			case key.Matches(msg, breakKeys.Quit):
				return m, tea.Quit
			}
		}

	case promptMsg:
		m.ready(string(msg))
		return m, nil

	case spinner.TickMsg:
		if m.state != StateLoading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI interface.
func (m model) View() string {
	return renderApp(m)
}

// Launch starts the TUI with the given text prompt as a scrollable typing
// interface.
func Launch(dirPath string) {
	p := tea.NewProgram(initialModel(dirPath))
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
