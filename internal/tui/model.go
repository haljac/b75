package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haljac/b75/internal/data"
	"github.com/haljac/b75/internal/runner"
	"github.com/haljac/b75/internal/tutor"
	"github.com/haljac/b75/internal/workspace"
)

func debugLog(format string, args ...interface{}) {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, time.Now().Format(time.RFC3339)+": "+format+"\n", args...)
}

var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	outputStyle   = lipgloss.NewStyle().Margin(1, 2).Border(lipgloss.NormalBorder()).Padding(0, 1)
	statusMessage = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			MarginTop(1)
)

type sessionState int

const (
	listView sessionState = iota
	outputView
	tutorView
)

type item struct {
	title  string
	desc   string
	slug   string
	status string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type testResultMsg runner.Result
type tutorMsg string
type errMsg string

type Model struct {
	list     list.Model
	viewport viewport.Model
	state    sessionState
	quitting bool
	ready    bool

	// Tutor state
	tutorClient *tutor.Client
	lastOutput  string // Store last test output for context
}

func NewModel() Model {
	items := []list.Item{}
	for _, p := range data.Problems {
		items = append(items, item{
			title:  p.Title,
			desc:   p.Description,
			slug:   p.Slug,
			status: "TODO", // TODO: Load from state
		})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Blind 75 Problems"
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
			key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test")),
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "tutor")),
		}
	}

	return Model{
		list:  l,
		state: listView,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == outputView || m.state == tutorView {
			if msg.String() == "esc" || msg.String() == "q" {
				m.state = listView
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		// List View Keys
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter", "e":
			i, ok := m.list.SelectedItem().(item)
			debugLog("Enter pressed. Selected: %v, OK: %v", i.title, ok)
			if ok {
				return m, openEditor(i.slug)
			} else {
				debugLog("Type assertion failed for selected item")
			}
		case "t":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				return m, runTests(i.slug)
			}
		case "?":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.state = tutorView
				m.viewport.SetContent("Consulting the oracle...")
				return m, askTutor(i.slug, i.title, i.desc, m.lastOutput)
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

		if !m.ready {
			m.viewport = viewport.New(msg.Width-h, msg.Height-v-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - h
			m.viewport.Height = msg.Height - v - 4
		}

	case testResultMsg:
		m.state = outputView
		m.lastOutput = msg.Output
		content := fmt.Sprintf("Passed: %v\n\n%s", msg.Passed, msg.Output)
		m.viewport.SetContent(content)
		return m, nil

	case tutorMsg:
		m.viewport.SetContent(string(msg))
		return m, nil

	case errMsg:
		m.state = outputView
		m.viewport.SetContent(fmt.Sprintf("Error:\n\n%v", string(msg)))
		return m, nil
	}

	if m.state == listView {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.state == outputView || m.state == tutorView {
		return docStyle.Render(m.viewport.View())
	}

	return docStyle.Render(m.list.View())
}

func openEditor(slug string) tea.Cmd {
	return func() tea.Msg {
		debugLog("openEditor called for %s", slug)
		if err := workspace.EnsureProblem(slug); err != nil {
			debugLog("EnsureProblem failed: %v", err)
			return errMsg(fmt.Sprintf("Failed to setup problem: %v", err))
		}

		path, err := workspace.GetProblemPath(slug)
		if err != nil {
			debugLog("GetProblemPath failed: %v", err)
			return errMsg(fmt.Sprintf("Failed to get path: %v", err))
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		debugLog("Using editor: %s", editor)

		if _, err := exec.LookPath(editor); err != nil {
			debugLog("LookPath failed: %v", err)
			return errMsg(fmt.Sprintf("Editor '%s' not found. Please set $EDITOR or install vim.", editor))
		}

		debugLog("Starting editor process at %s/main.go", path)
		c := exec.Command(editor, filepath.Join(path, "main.go"))
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				debugLog("ExecProcess finished with error: %v", err)
				return errMsg(fmt.Sprintf("Editor exited with error: %v", err))
			}
			debugLog("ExecProcess finished successfully")
			return nil
		})
	}
}

func runTests(slug string) tea.Cmd {
	return func() tea.Msg {
		if err := workspace.EnsureProblem(slug); err != nil {
			return testResultMsg{Passed: false, Output: err.Error()}
		}

		res, err := runner.RunTests(slug)
		if err != nil {
			return testResultMsg{Passed: false, Output: err.Error()}
		}
		return testResultMsg(res)
	}
}

func askTutor(slug, title, desc, testOutput string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := tutor.NewClient(ctx)
		if err != nil {
			return tutorMsg("Error initializing Gemini: " + err.Error())
		}

		path, err := workspace.GetProblemPath(slug)
		if err != nil {
			return tutorMsg("Error finding problem: " + err.Error())
		}

		codeBytes, err := os.ReadFile(filepath.Join(path, "main.go"))
		if err != nil {
			return tutorMsg("Error reading code: " + err.Error())
		}

		resp, err := client.Ask(ctx, title, desc, string(codeBytes), testOutput)
		if err != nil {
			return tutorMsg("Error from Gemini: " + err.Error())
		}

		return tutorMsg(resp)
	}
}
