package components

import (
    "fmt"
    "strings"


    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/MehdiShekari/go_tui/models"
)

// TaskListModel displays a scrollable list of tasks with keyboard navigation
type TaskListModel struct {
    tasks    []models.Task
    cursor   int
    viewport viewport.Model
    width    int
    height   int
    selected int64
    styles   TaskListStyles
}

// TaskListStyles defines the styling options for the task list
type TaskListStyles struct {
    Normal      lipgloss.Style
    Selected    lipgloss.Style
    Priority    map[models.Priority]lipgloss.Style
    Overdue     lipgloss.Style
    DueToday    lipgloss.Style
    Tag         lipgloss.Style
    Status      map[models.Status]lipgloss.Style
}

// DefaultTaskListStyles returns the default styling for task lists
func DefaultTaskListStyles() TaskListStyles {
    return TaskListStyles{
        Normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
        Selected: lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("255")),
        Priority: map[models.Priority]lipgloss.Style{
            models.PriorityLow:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
            models.PriorityMedium: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00")),
            models.PriorityHigh:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8800")),
            models.PriorityUrgent: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
        },
        Overdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
        DueToday: lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
        Tag:      lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Background(lipgloss.Color("235")),
        Status: map[models.Status]lipgloss.Style{
            models.StatusTodo:       lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
            models.StatusInProgress: lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
            models.StatusDone:       lipgloss.NewStyle().Foreground(lipgloss.Color("78")),
            models.StatusArchived:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        },
    }
}

// NewTaskListModel creates a new task list component
func NewTaskListModel() TaskListModel {
    return TaskListModel{
        viewport: viewport.New(0, 0),
        styles:   DefaultTaskListStyles(),
    }
}

// Init initializes the task list model
func (m TaskListModel) Init() tea.Cmd {
    return nil
}

// Update handles messages and updates the task list state
func (m TaskListModel) Update(msg tea.Msg) (TaskListModel, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
                m.updateViewport()
            }

        case "down", "j":
            if m.cursor < len(m.tasks)-1 {
                m.cursor++
                m.updateViewport()
            }

        case "home", "g":
            m.cursor = 0
            m.viewport.GotoTop()

        case "end", "G":
            if len(m.tasks) > 0 {
                m.cursor = len(m.tasks) - 1
                m.viewport.GotoBottom()
            }

        case "enter":
            if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
                m.selected = m.tasks[m.cursor].ID
            }

        case "pgup":
            m.cursor = max(0, m.cursor-10)
            m.viewport.LineUp(10)

        case "pgdown":
            m.cursor = min(len(m.tasks)-1, m.cursor+10)
            m.viewport.LineDown(10)
        }
    }

    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

// View renders the task list
func (m TaskListModel) View() string {
    if len(m.tasks) == 0 {
        emptyStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            Italic(true).
            Padding(2).
            Align(lipgloss.Center)
        return emptyStyle.Render("No tasks found. Press 'n' to create a new task.")
    }

    var lines []string
    for i, task := range m.tasks {
        line := m.renderTaskLine(i, task)
        lines = append(lines, line)
    }

    content := strings.Join(lines, "\n")
    m.viewport.SetContent(content)

    return lipgloss.NewStyle().
        Padding(0, 1).
        Render(m.viewport.View())
}

// renderTaskLine renders a single task line with formatting
func (m TaskListModel) renderTaskLine(index int, task models.Task) string {
    // Cursor indicator
    cursor := "  "
    if m.cursor == index {
        cursor = lipgloss.NewStyle().
            Foreground(lipgloss.Color("205")).
            Render("> ")
    }

    // Status icon
    statusIcon := task.Status.Icon()
    statusStyle := m.styles.Status[task.Status]
    status := statusStyle.Render(statusIcon)

    // Priority indicator
    priorityStyle := m.styles.Priority[task.Priority]
    priority := priorityStyle.Render(fmt.Sprintf("[%s]", strings.ToUpper(task.Priority.String()[:1])))

    // Task title with truncation
    titleMaxLen := m.width - 40
    title := task.Title
    if len(title) > titleMaxLen && titleMaxLen > 3 {
        title = title[:titleMaxLen-3] + "..."
    }

    // Apply completion styling
    if task.Status == models.StatusDone || task.Status == models.StatusArchived {
        title = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            Strikethrough(true).
            Render(title)
    }

    // Due date
    dueDate := m.formatDueDate(task)

    // Tags
    tags := m.formatTags(task.Tags)

    // Assemble line
    line := fmt.Sprintf("%s%s %s %s %s%s",
        cursor, status, priority, title, dueDate, tags,
    )

    // Apply selection styling
    if m.cursor == index {
        return m.styles.Selected.Render(line)
    }

    return m.styles.Normal.Render(line)
}

// formatDueDate formats the due date with appropriate styling
func (m TaskListModel) formatDueDate(task models.Task) string {
    if task.DueDate == nil {
        return ""
    }

    if task.IsOverdue() {
        return " " + m.styles.Overdue.Render("OVERDUE: "+task.DueDate.Format("2006-01-02"))
    }

    if task.IsDueToday() {
        return " " + m.styles.DueToday.Render("TODAY")
    }

    return " " + lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")).
        Render(task.DueDate.Format("2006-01-02"))
}

// formatTags formats tags with styling
func (m TaskListModel) formatTags(tags []string) string {
    if len(tags) == 0 {
        return ""
    }

    formatted := make([]string, len(tags))
    for i, tag := range tags {
        formatted[i] = m.styles.Tag.Render(" " + tag + " ")
    }

    return " " + strings.Join(formatted, " ")
}

// updateViewport adjusts the viewport to keep the cursor visible
func (m *TaskListModel) updateViewport() {
    if m.cursor < m.viewport.YOffset {
        m.viewport.SetYOffset(m.cursor)
    } else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
        m.viewport.SetYOffset(m.cursor - m.viewport.Height + 1)
    }
}

// SetTasks updates the task list with new tasks
func (m *TaskListModel) SetTasks(tasks []models.Task) {
    m.tasks = tasks
    if m.cursor >= len(tasks) {
        m.cursor = max(0, len(tasks)-1)
    }
}

// SetSize sets the dimensions of the task list
func (m *TaskListModel) SetSize(width, height int) {
    m.width = width
    m.height = height
    m.viewport.Width = width
    m.viewport.Height = height
}

// GetSelected returns the ID of the selected task, or 0 if none selected
func (m *TaskListModel) GetSelected() int64 {
    return m.selected
}

// ClearSelected resets the selected task ID
func (m *TaskListModel) ClearSelected() {
    m.selected = 0
}

// GetTasks returns all tasks currently in the list
func (m *TaskListModel) GetTasks() []models.Task {
    return m.tasks
}

// Cursor returns the current cursor position
func (m *TaskListModel) Cursor() int {
    return m.cursor
}

// SelectedTask returns the currently selected task, or nil if none
func (m *TaskListModel) SelectedTask() *models.Task {
    if m.cursor < 0 || m.cursor >= len(m.tasks) {
        return nil
    }
    return &m.tasks[m.cursor]
}

// Helper functions
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}