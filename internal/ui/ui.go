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

	// Domain-specific
	promptPoolSize = 3
	maxPromptLen   = 128
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
	// mistakes records the positions of mistakes made.
	mistakes map[int]bool

	// promptPool holds pre-fetched prompts.
	promptPool []string
	// promptsBeingFetched tracks the number of ongoing prompt fetches.
	promptsBeingFetched int

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

// cursor returns the current cursor position within the prompt.
func (m model) cursor() int {
	return len([]rune(m.input))
}

// promptFetchedMsg is a message indicating a prompt has been fetched.
type promptFetchedMsg string

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
		spinner: spinner,
	}

	return m
}

// load sets up the model for a loading state.
func (m model) load() model {
	m.state = StateLoading
	return m
}

// ready sets up the model for a ready state with a new prompt.
func (m model) ready() model {
	m.mistakes = make(map[int]bool)
	m.input = ""
	m.wpm = 0
	m.accuracy = 0.0
	m.state = StateReady
	return m
}

// start begins the typing session.
func (m model) start() model {
	m.state = StateSession
	m.startTime = time.Now()
	return m
}

// stop ends the typing session.
func (m model) stop() model {
	m.state = StateBreak
	return m
}

// initMsg is the initial message to start the TUI.
type initMsg struct{}

// initCmd returns the initial command to start the TUI.
func initCmd() tea.Cmd {
	return func() tea.Msg {
		return initMsg{}
	}
}

// fetchMorePromptsIfNeeded initiates fetching another prompt asynchronously.
// If a fetch is already in progress or the pool is sufficiently full, it does
// nothing.
// Make sure to capture the new model state to reflect the necessary changes.
func (m model) fetchMorePromptsIfNeeded() (model, tea.Cmd) {
	if len(m.promptPool) >= promptPoolSize || m.promptsBeingFetched > 0 {
		return m, nil
	}

	m.promptsBeingFetched++
	return m, func() tea.Msg {
		prompt, err := domain.Prompt(m.dirPath, maxPromptLen)
		if err != nil {
			panic(err) // TODO: graceful error handling
		}
		return promptFetchedMsg(prompt)
	}
}

// acceptPrompt adds a fetched prompt to the model's pool.
func (m model) acceptPrompt(prompt string) model {
	m.promptsBeingFetched--
	m.promptPool = append(m.promptPool, prompt)
	return m
}

// consumePrompt removes the next prompt from the pool and sets it as the
// current prompt.
func (m model) consumePrompt() model {
	if len(m.promptPool) == 0 {
		return m
	}
	prompt := m.promptPool[0]
	m.promptPool = m.promptPool[1:]
	m.prompt = prompt
	return m
}

// Init is the initial command for the TUI.
func (m model) Init() tea.Cmd {
	return initCmd()
}

// handleBackspace processes a backspace key press.
func (m model) handleBackspace() model {
	if m.cursor() == 0 {
		return m
	}

	inputRunes := []rune(m.input)
	m.input = string(inputRunes[:len(inputRunes) - 1])
	return m
}

// handleCtrlBackspace processes a Ctrl+Backspace key press.
func (m model) handleCtrlBackspace() model {
	inputRunes := []rune(m.input)
	promptRunes := []rune(m.prompt)

	cursor := m.cursor()
	for cursor > 0 && (promptRunes[cursor-1] != ' ' || cursor == m.cursor()) {
		cursor--
	}

	m.input = string(inputRunes[:cursor])
	return m
}

// Update handles incoming messages and updates the TUI state.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case initMsg:
		var cmd tea.Cmd
		m, cmd = m.load().fetchMorePromptsIfNeeded()
		return m, tea.Batch(m.spinner.Tick, cmd)

	case tea.KeyMsg:
		// Quit if Ctrl+C is pressed
		if key.Matches(msg, globalKeys.Quit) {
			return m, tea.Quit
		}

		if m.state == StateSession || m.state == StateReady {
			if key.Matches(msg, sessionKeys.Stop) {
				return m.stop(), nil
			}

			switch msg.String() {
			case "backspace":
				return m.handleBackspace(), nil
			case "ctrl+backspace", "ctrl+w":
				return m.handleCtrlBackspace(), nil
			default:
				msgRunes := []rune(msg.String())
				promptRunes := []rune(m.prompt)

				// Ignore non-character keys or unsupported runes
				if len(msgRunes) != 1 ||
					!slices.Contains(allowedInputRunes, msgRunes[0]) {
					return m, nil
				}

				// Start session on first valid input
				if m.state == StateReady {
					m = m.start()
				}

				// Check for mistake
				if msg.String() != string(promptRunes[m.cursor()]) {
					m.mistakes[m.cursor()] = true
				}

				// Accept input
				m.input += msg.String()

				// Update metrics
				elapsed := time.Since(m.startTime)
				m.wpm = int(math.Round(metrics.WPM(m.input, elapsed)))
				m.accuracy = int(
					math.Round(metrics.Accuracy(m.prompt, m.input)))

				// End session if prompt completed
				if m.cursor() >= len(promptRunes) {
					return m.stop(), nil
				}
			}
		} else {
			switch {
			case key.Matches(msg, breakKeys.Restart):
				// Check if more prompts are available
				if len(m.promptPool) == 0 {
					var cmd tea.Cmd
					m, cmd = m.load().fetchMorePromptsIfNeeded()
					return m, tea.Batch(cmd, m.spinner.Tick)
				}

				// Load next prompt from pool
				m = m.consumePrompt().ready()

				return m.fetchMorePromptsIfNeeded()

			case key.Matches(msg, breakKeys.Quit):
				return m, tea.Quit
			}
		}

	case promptFetchedMsg:
		m = m.acceptPrompt(string(msg))

		if m.state == StateLoading {
			m = m.consumePrompt().ready()
		}

		// Fetch more prompts if pool is low
		return m.fetchMorePromptsIfNeeded()

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
