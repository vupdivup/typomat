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
	"github.com/vupdivup/typomat/internal/domain"
	"github.com/vupdivup/typomat/pkg/alphabet"
	"github.com/vupdivup/typomat/pkg/metrics"
	"go.uber.org/zap"
)

const (
	// windowOuterWidth is the total width of the TUI, including borders.
	windowOuterWidth = 80
	// windowPaddedWidth is the width of the window including padding but no
	// borders.
	windowPaddedWidth = windowOuterWidth - 2
	// windowPaddingVertical defines the vertical padding inside the window.
	windowPaddingVertical = 0
	// windowPaddingHorizontal defines the horizontal padding inside the window.
	windowPaddingHorizontal = 1
	// windowContentWidth is the width available for content inside the window.
	windowContentWidth = windowOuterWidth - 2 - 2*windowPaddingHorizontal

	// canvasPaddingVertical defines the vertical padding inside the canvas.
	canvasPaddingVertical = 1
	// canvasPaddingHorizontal defines the horizontal padding inside the canvas.
	canvasPaddingHorizontal = 3
	// canvasContentWidth is the width available for content inside the canvas.
	canvasContentWidth = windowContentWidth - 2*canvasPaddingHorizontal

	// promptPoolSize is the number of pre-fetched prompts to maintain.
	promptPoolSize = 3
	// maxPromptLen is the maximum length of a typing prompt.
	maxPromptLen = 128
)

var (
	// Styles
	windowStyle = lipgloss.NewStyle().
			Width(windowPaddedWidth).
			Padding(windowPaddingVertical, windowPaddingHorizontal).
			Border(lipgloss.RoundedBorder()).
			BorderTop(false).
			BorderForeground(mutedColor)

	canvasStyle = lipgloss.NewStyle().
			Padding(canvasPaddingVertical, canvasPaddingHorizontal).
			Height(4)

	accentColor = lipgloss.Color("3")
	bodyColor   = lipgloss.Color("7")
	mutedColor  = lipgloss.Color("8")
	errorColor  = lipgloss.Color("9")

	accentStyle = lipgloss.NewStyle().Foreground(accentColor)
	bodyStyle   = lipgloss.NewStyle().Foreground(bodyColor)
	mutedStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	errorStyle  = lipgloss.NewStyle().Foreground(errorColor)

	allowedInputRunes = append(alphabet.AllRunes, ' ')
)

// AppState represents the current state of the application.
type AppState int

const (
	// StateLoading indicates the application is loading resources.
	StateLoading AppState = iota
	// StateSession indicates an active typing session.
	StateSession
	// StateBreak indicates a break between typing sessions.
	StateBreak
	// StateReady indicates the application is ready for a new typing session.
	StateReady
)

// model defines the TUI state.
type model struct {
	// dirPath is the directory path for prompts.
	dirPath string

	// appState is the current application appState.
	appState AppState

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

	// Err captures any error that occurs during TUI execution.
	err error

	// frameTime is the time of the last frame update.
	frameTime time.Time
}

// cursor returns the current cursor position within the prompt.
func (m model) cursor() int {
	return len([]rune(m.input))
}

// promptFetchedMsg is a message indicating a prompt has been fetched.
type promptFetchedMsg struct {
	prompt string
	err    error
}

// initialModel creates the initial TUI model.
func initialModel(dirPath string) model {
	help := help.New()
	help.Styles.ShortKey = accentStyle
	help.Styles.ShortSeparator = mutedStyle

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
	m.appState = StateLoading
	return m
}

// ready sets up the model for a ready state with a new prompt.
func (m model) ready() model {
	m.mistakes = make(map[int]bool)
	m.input = ""
	m.wpm = 0
	m.accuracy = 0.0
	m.appState = StateReady
	zap.S().Infow("Session ready",
		"prompt", m.prompt)
	return m
}

// start begins the typing session.
func (m model) start() model {
	m.appState = StateSession
	m.startTime = m.frameTime
	return m
}

// stop ends the typing session.
func (m model) stop() model {
	m.appState = StateBreak
	zap.S().Infow("Session ended",
		"prompt", m.prompt,
		"input", m.input,
		"wpm", m.wpm,
		"accuracy", m.accuracy)
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
		return promptFetchedMsg{prompt: prompt, err: err}
	}
}

// acceptPrompt adds a fetched prompt to the model's pool.
func (m model) acceptPrompt(prompt string) model {
	m.promptsBeingFetched--
	m.promptPool = append(m.promptPool, prompt)
	zap.S().Debugw("Pooled new prompt",
		"prompt", prompt,
		"pool_size", len(m.promptPool))
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
	zap.S().Debugw("Consumed prompt from pool",
		"prompt", prompt,
		"pool_size", len(m.promptPool))
	return m
}

// Init is the initial command for the TUI.
func (m model) Init() tea.Cmd {
	return initCmd()
}

func (m model) updateMetrics() model {
	elapsed := m.frameTime.Sub(m.startTime)
	m.wpm = int(math.Round(metrics.WPM(m.input, elapsed)))
	m.accuracy = int(
		math.Round(metrics.Accuracy(m.prompt, m.input)))
	return m
}

// handleBackspace processes a backspace key press.
// Updates metrics as well.
func (m model) handleBackspace() model {
	if m.cursor() == 0 {
		return m
	}

	inputRunes := []rune(m.input)
	m.input = string(inputRunes[:len(inputRunes)-1])
	m = m.updateMetrics()
	zap.S().Debugw("Handled backspace",
		"input", m.input,
		"wpm", m.wpm,
		"accuracy", m.accuracy)
	return m
}

// handleCtrlBackspace processes a Ctrl+Backspace key press.
// Updates metrics as well.
func (m model) handleCtrlBackspace() model {
	inputRunes := []rune(m.input)
	promptRunes := []rune(m.prompt)

	cursor := m.cursor()
	for cursor > 0 && (promptRunes[cursor-1] != ' ' || cursor == m.cursor()) {
		cursor--
	}

	m.input = string(inputRunes[:cursor])
	m = m.updateMetrics()
	zap.S().Debugw("Handled ctrl+backspace",
		"input", m.input,
		"wpm", m.wpm,
		"accuracy", m.accuracy)

	return m
}

// Update handles incoming messages and updates the TUI state.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.frameTime = time.Now()

	switch msg := msg.(type) {

	case initMsg:
		var cmd tea.Cmd
		m, cmd = m.load().fetchMorePromptsIfNeeded()
		return m, tea.Batch(m.spinner.Tick, cmd)

	case tea.KeyMsg:
		if key.Matches(msg, globalKeys.Quit) {
			return m, tea.Quit
		}

		switch m.appState {
		case StateBreak:
			if key.Matches(msg, breakKeys.Restart) {
				// Check if more prompts are available
				if len(m.promptPool) == 0 {
					var cmd tea.Cmd
					m, cmd = m.load().fetchMorePromptsIfNeeded()
					return m, tea.Batch(cmd, m.spinner.Tick)
				}

				// Load next prompt from pool
				m = m.consumePrompt().ready()

				return m.fetchMorePromptsIfNeeded()
			}

		case StateSession, StateReady:
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
				if m.appState == StateReady {
					m = m.start()
				}

				// Check for mistake
				if msg.String() != string(promptRunes[m.cursor()]) {
					m.mistakes[m.cursor()] = true
				}

				// Accept input
				m.input += msg.String()

				// Update metrics
				m = m.updateMetrics()

				// End session if prompt completed
				if m.cursor() >= len(promptRunes) {
					return m.stop(), nil
				}
			}
			zap.S().Debugw("Handled key press",
				"msg", msg.String(),
				"input", m.input,
				"wpm", m.wpm,
				"accuracy", m.accuracy)
			return m, nil
		}

	case promptFetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}

		m = m.acceptPrompt(msg.prompt)

		if m.appState == StateLoading {
			m = m.consumePrompt().ready()
		}

		// Fetch more prompts if pool is low
		return m.fetchMorePromptsIfNeeded()

	case spinner.TickMsg:
		if m.appState != StateLoading {
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
	// If there was an error, exit without rendering the app.
	if m.err != nil {
		return "\n"
	}

	return renderApp(m)
}

// Launch starts the TUI with the given text prompt as a scrollable typing
// interface.
func Launch(dirPath string) error {
	p := tea.NewProgram(initialModel(dirPath))
	m, err := p.Run()
	if err != nil {
		return err
	}

	uiModel := m.(model)
	return uiModel.err
}
