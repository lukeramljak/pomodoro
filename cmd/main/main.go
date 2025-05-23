package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const workTime time.Duration = time.Minute * 25
const breakTime time.Duration = time.Minute * 5

var workSpinner = spinner.Jump
var breakSpinner = spinner.Dot

type model struct {
	spinner        spinner.Model
	timer          timer.Model
	keymap         keymap
	help           help.Model
	quitting       bool
	isBreak        bool
	containerStyle lipgloss.Style
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.timer.Init(),
		m.spinner.Tick,
	)
}

func (m model) View() string {
	var s string

	s = m.spinner.View()

	if m.isBreak {
		s += "Break ends in " + m.timer.View()
	} else {
		s += "Work ends in " + m.timer.View()
	}

	s += "\n\n"
	s += m.helpView()

	return m.containerStyle.Render(s)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		if m.isBreak {
			notify("Back to work!", "Sorry!")
			playSound("Submarine")

			m.isBreak = false
			m.timer.Timeout = workTime
			m.spinner.Spinner = workSpinner
		} else {
			notify("Break time!", "Time to stretch and relax")
			playSound("Glass")

			m.isBreak = true
			m.timer.Timeout = breakTime
			m.spinner.Spinner = breakSpinner
		}

		return m, tea.Batch(m.timer.Start(), m.spinner.Tick)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.isBreak = false
			m.spinner.Spinner = workSpinner
			m.timer.Timeout = workTime
			return m, tea.Batch(m.timer.Start(), m.spinner.Tick)
		case key.Matches(msg, m.keymap.start, m.keymap.stop):
			return m, m.timer.Toggle()
		}
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func main() {
	s := spinner.New()
	s.Spinner = workSpinner
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).PaddingRight(2)

	containerStyle := lipgloss.NewStyle().Padding(2)

	m := model{
		spinner: s,
		timer:   timer.NewWithInterval(workTime, time.Second),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
		help:           help.New(),
		containerStyle: containerStyle,
	}
	m.keymap.start.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Uh oh, we encountered an error:", err)
		os.Exit(1)
	}
}

func notify(title, message string) {
	cmd := exec.Command("osascript", "-e",
		`display notification "`+message+`" with title "`+title+`"`,
	)
	_ = cmd.Run()
}

func playSound(name string) {
	path := "/System/Library/Sounds/" + name + ".aiff"
	_ = exec.Command("afplay", path).Run()
}
