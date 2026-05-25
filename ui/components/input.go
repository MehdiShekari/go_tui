package components

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// InputModel represents a form with multiple input fields
type InputModel struct {
    inputs   []textinput.Model
    focused  int
    title    string
    width    int
    height   int
    done     bool
    canceled bool
}

// NewInputModel creates a new input form with the specified title and field placeholders
func NewInputModel(title string, fields []string) InputModel {
    m := InputModel{
        inputs:  make([]textinput.Model, len(fields)),
        title:   title,
        focused: 0,
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

// Init initializes the input model
func (m InputModel) Init() tea.Cmd {
    return textinput.Blink
}

// Update handles messages and updates the input state
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            m.canceled = true
            m.done = true
            return m, nil

        case "enter":
            // Submit the form
            m.done = true
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

// nextField moves focus to the next input field
func (m *InputModel) nextField() {
    m.focused = (m.focused + 1) % len(m.inputs)

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

// prevField moves focus to the previous input field
func (m *InputModel) prevField() {
    m.focused--
    if m.focused < 0 {
        m.focused = len(m.inputs) - 1
    }

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

// updateInputs updates all input fields with the given message
func (m *InputModel) updateInputs(msg tea.Msg) tea.Cmd {
    cmds := make([]tea.Cmd, len(m.inputs))

    for i := range m.inputs {
        m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
    }

    return tea.Batch(cmds...)
}

// View renders the input form
func (m InputModel) View() string {
    var s strings.Builder

    // Title
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205")).
        MarginBottom(1)
    s.WriteString(titleStyle.Render(m.title) + "\n\n")

    // Input fields
    for i := range m.inputs {
        s.WriteString(m.inputs[i].View())
        if i < len(m.inputs)-1 {
            s.WriteString("\n\n")
        }
    }

    // Help text
    s.WriteString("\n\n")
    helpStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))
    s.WriteString(helpStyle.Render("Tab/Shift+Tab: Navigate fields | Enter: Submit | Esc: Cancel"))

    // Border and padding
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("63")).
        Padding(1, 2).
        Width(m.width - 4).
        Render(s.String())
}

// Values returns the current values of all input fields
func (m InputModel) Values() []string {
    values := make([]string, len(m.inputs))
    for i, input := range m.inputs {
        values[i] = input.Value()
    }
    return values
}

// Done returns whether the form has been submitted or canceled
func (m InputModel) Done() bool {
    return m.done
}

// Canceled returns whether the form was canceled
func (m InputModel) Canceled() bool {
    return m.canceled
}

// SetSize sets the dimensions of the input form
func (m *InputModel) SetSize(width, height int) {
    m.width = width
    m.height = height
}

// SetValues pre-fills the input fields with the given values
func (m *InputModel) SetValues(values []string) {
    for i, val := range values {
        if i < len(m.inputs) {
            m.inputs[i].SetValue(val)
        }
    }
}

// FocusedIndex returns the index of the currently focused input field
func (m InputModel) FocusedIndex() int {
    return m.focused
}

// String returns a summary of the form values
func (m InputModel) String() string {
    var parts []string
    for i, input := range m.inputs {
        parts = append(parts, fmt.Sprintf("Field %d: %s", i+1, input.Value()))
    }
    return strings.Join(parts, "\n")
}