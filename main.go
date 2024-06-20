package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	kwStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
	outterStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 0, 0, 3)
)

func main() {
	var data []byte
	var err error

	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	p := tea.NewProgram(initialModel(data), tea.WithAltScreen())

	if m, err := p.Run(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(m.(model).regex.Value())
	}
}

type errMsg error

type model struct {
	regex  textinput.Model
	text   textarea.Model
	out    string
	rx     *regexp.Regexp
	width  int
	height int
	err    error
}

func initialModel(input []byte) model {
	ti := textinput.New()
	ti.Placeholder = "regex here"

	rx, _ := regexp.Compile(".")
	text := textarea.New()
	text.SetValue(string(input))

	text.Focus()
	return model{
		regex: ti,
		rx:    rx,
		text:  text,
		err:   nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			if m.regex.Focused() {
				cmd = m.text.Focus()
				m.regex.Blur()
			} else {
				cmd = m.regex.Focus()
				m.text.Blur()
			}
			cmds = append(cmds, cmd)
		default:
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.sizeInputs()
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.regex, cmd = m.regex.Update(msg)
	cmds = append(cmds, cmd)
	m.text, cmd = m.text.Update(msg)
	cmds = append(cmds, cmd)

	rx, err := regexp.Compile(m.regex.Value())
	m.rx = rx
	m.err = err
	if err != nil {
		return m, tea.Batch(cmds...)
	}
	txt := m.text.Value()
	idxs := rx.FindAllStringIndex(txt, -1)
	_ = idxs
	i := 0
	currI := 0
	out := ""
	for len(idxs) > 0 {
		if currI == len(idxs) {
			out += txt[idxs[currI-1][1]:]
			break
		}
		if i <= idxs[currI][0] {
			out += txt[i:idxs[currI][0]]
			out += kwStyle.Render(txt[idxs[currI][0]:idxs[currI][1]])
			i = idxs[currI][1]
			currI++
			continue
		} else {
			break
		}
	}
	if len(idxs) == 0 {
		out = txt
	}
	m.out = out
	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	m.text.SetWidth(m.width)
	m.text.SetHeight(m.height/2 - 2)
	outterStyle = outterStyle.Width(m.width - 2).Height(m.height/2 - 2)
}

func (m model) View() string {
	err := ""
	if m.err != nil {
		err = m.err.Error()
	}
	return fmt.Sprintf(
		"    %s\n%s\n    %s\n%s",
		m.regex.View(),
		m.text.View(),
		kwStyle.Render(err),
		outterStyle.Render(m.out),
	)
}
