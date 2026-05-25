package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel displays a configurable form of multiple text input fields
type InputModel struct {
	inputs   []textinput.Model
	fields   []string
	title    string
	focused  int
	done     bool
	canceled bool
	width    int
	height   int
	errs     map[int]string
}

// NewInputModel initializes the form input controller with field labels
func NewInputModel(title string, fields []string) InputModel {
	m := InputModel{
		inputs:  make([]textinput.Model, len(fields)),
		fields:  fields,
		title:   title,
		focused: 0,
		errs:    make(map[int]string),
	}

	for i, field := range fields {
		t := textinput.New()
		t.Placeholder = field
		t.CharLimit = 200
		t.Width = 60

		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}
		m.inputs[i] = t
	}

	return m
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			m.done = true
			return m, nil

		case "enter":
			// Fix: Enter only saves when pressing it on the final field.
			// Otherwise, it moves down to the next text field.
			if m.focused == len(m.inputs)-1 || len(m.inputs) == 1 {
				m.done = true
				return m, nil
			}
			m.nextField()
			return m, nil

		case "tab":
			m.nextField()
		case "shift+tab":
			m.prevField()
		case "up":
			m.prevField()
		case "down":
			m.nextField()
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *InputModel) nextField() {
	if len(m.inputs) > 0 {
		m.focused = (m.focused + 1) % len(m.inputs)
		m.refreshFocus()
	}
}

func (m *InputModel) prevField() {
	if len(m.inputs) > 0 {
		m.focused--
		if m.focused < 0 {
			m.focused = len(m.inputs) - 1
		}
		m.refreshFocus()
	}
}

func (m *InputModel) refreshFocus() {
	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		}
	}
}

func (m *InputModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m InputModel) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(m.title) + "\n\n")

	for i, input := range m.inputs {
		if i >= len(m.fields) {
			break
		}
		
		b.WriteString(fmt.Sprintf("%s:\n", m.fields[i]))
		b.WriteString(input.View() + "\n")

		if errMsg, exists := m.errs[i]; exists && errMsg != "" {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Italic(true)
			b.WriteString(errStyle.Render("   ↳ "+errMsg) + "\n")
		}
		b.WriteString("\n")
	}

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	b.WriteString(footerStyle.Render("\n[Tab/Down] Next Field • [Enter] Next Field or Submit • [Esc] Discard"))
	return b.String()
}

func (m InputModel) GetValues() []string {
	values := make([]string, len(m.inputs))
	for i, input := range m.inputs {
		values[i] = input.Value()
	}
	return values
}

func (m InputModel) Title() string {
	return m.title
}

func (m *InputModel) SetFieldError(index int, errMsg string) {
	m.errs[index] = errMsg
}

func (m *InputModel) ClearErrors() {
	m.errs = make(map[int]string)
}

func (m *InputModel) ResetDoneStatus() {
	m.done = false
}

func (m InputModel) Done() bool {
	return m.done
}

func (m InputModel) Canceled() bool {
	return m.canceled
}

func (m *InputModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *InputModel) SetValues(values []string) {
	for i, val := range values {
		if i < len(m.inputs) {
			m.inputs[i].SetValue(val)
		}
	}
}

func (m InputModel) FocusedIndex() int {
	return m.focused
}

func (m InputModel) String() string {
	var parts []string
	for i, input := range m.inputs {
		parts = append(parts, fmt.Sprintf("Field %d: %s", i+1, input.Value()))
	}
	return strings.Join(parts, "\n")
}