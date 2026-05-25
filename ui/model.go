package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "github.com/MehdiShekari/go_tui/db"
    "github.com/MehdiShekari/go_tui/models"
    "github.com/MehdiShekari/go_tui/ui/components"
)

// View represents the different screens in the application
type View int

const (
    mainView View = iota
    detailView
    inputView
    helpView
    statsView
    filterView
    confirmView
)

// keyMap defines all keyboard shortcuts for the application
type keyMap struct {
    Up        key.Binding
    Down      key.Binding
    Enter     key.Binding
    New       key.Binding
    Edit      key.Binding
    Delete    key.Binding
    Quit      key.Binding
    Filter    key.Binding
    Help      key.Binding
    Stats     key.Binding
    Back      key.Binding
    Priority  key.Binding
    Status    key.Binding
    Search    key.Binding
    ClearFilter key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.Enter},
        {k.New, k.Edit, k.Delete},
        {k.Filter, k.Search, k.ClearFilter},
        {k.Priority, k.Status, k.Stats},
        {k.Help, k.Quit},
    }
}

var keys = keyMap{
    Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "move up")),
    Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "move down")),
    Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select/view")),
    New:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new task")),
    Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit task")),
    Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete task")),
    Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    Filter:      key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter tasks")),
    Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
    Stats:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "statistics")),
    Back:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
    Priority:    key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "change priority")),
    Status:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "change status")),
    Search:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
    ClearFilter: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "clear filters")),
}

// Model represents the main application model
type Model struct {
    db           *database.Database
    list         components.TaskListModel
    input        components.InputModel
    help         help.Model
    keys         keyMap
    currentView  View
    tasks        []models.Task
    filter       models.TaskFilter
    width        int
    height       int
    ready        bool
    err          error
    stats        map[string]int
    statusMsg    string
}

// NewModel creates a new application model
func NewModel(db *database.Database) Model {
    return Model{
        db:          db,
        list:        components.NewTaskListModel(),
        help:        help.New(),
        keys:        keys,
        currentView: mainView,
        filter:      models.TaskFilter{},
        stats:       make(map[string]int),
    }
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
    return m.loadTasksCmd()
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.help.Width = msg.Width
        m.list.SetSize(msg.Width, msg.Height-6)
        m.ready = true
        return m, nil

    case tea.KeyMsg:
        // Handle input view separately
        if m.currentView == inputView {
            return m.handleInputUpdate(msg)
        }

        // Handle confirmation view
        if m.currentView == confirmView {
            return m.handleConfirmUpdate(msg)
        }

        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit

        case key.Matches(msg, m.keys.Help):
            m.currentView = helpView
            return m, nil

        case key.Matches(msg, m.keys.Back):
            m.currentView = mainView
            m.statusMsg = ""
            return m, m.loadTasksCmd()

        case key.Matches(msg, m.keys.New):
            return m.showTaskForm("New Task", "", "", "")

        case key.Matches(msg, m.keys.Edit):
            task := m.list.SelectedTask()
            if task != nil {
                tags := strings.Join(task.Tags, ", ")
                return m.showTaskForm("Edit Task", task.Title, task.Description, tags)
            }

        case key.Matches(msg, m.keys.Delete):
            task := m.list.SelectedTask()
            if task != nil {
                m.currentView = confirmView
                m.statusMsg = fmt.Sprintf("Delete task: %s?", task.Title)
                return m, nil
            }

        case key.Matches(msg, m.keys.Search):
            return m.showInputPrompt("Search", "Enter search term")

        case key.Matches(msg, m.keys.Filter):
            m.currentView = filterView
            return m, nil

        case key.Matches(msg, m.keys.Stats):
            return m, m.loadStatsCmd()

        case key.Matches(msg, m.keys.ClearFilter):
            m.filter = models.TaskFilter{}
            m.statusMsg = "Filters cleared"
            return m, m.loadTasksCmd()

        case key.Matches(msg, m.keys.Priority):
            task := m.list.SelectedTask()
            if task != nil {
                return m, m.cycleTaskPriority(task.ID)
            }

        case key.Matches(msg, m.keys.Status):
            task := m.list.SelectedTask()
            if task != nil {
                return m, m.cycleTaskStatus(task.ID)
            }
        }

    case tasksLoadedMsg:
        m.tasks = msg.tasks
        m.list.SetTasks(msg.tasks)

    case taskCreatedMsg:
        m.statusMsg = fmt.Sprintf("Task created: %s", msg.task.Title)
        m.currentView = mainView
        cmds = append(cmds, m.loadTasksCmd())

    case taskUpdatedMsg:
        m.statusMsg = fmt.Sprintf("Task updated: %s", msg.task.Title)
        m.currentView = mainView
        cmds = append(cmds, m.loadTasksCmd())

    case taskDeletedMsg:
        m.statusMsg = "Task deleted"
        m.currentView = mainView
        cmds = append(cmds, m.loadTasksCmd())

    case statsLoadedMsg:
        m.stats = msg.stats
        m.currentView = statsView

    case errorMsg:
        m.err = msg.err
        m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
    }

    // Update the task list
    var listCmd tea.Cmd
    m.list, listCmd = m.list.Update(msg)
    cmds = append(cmds, listCmd)

    return m, tea.Batch(cmds...)
}

// View renders the application
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    if m.err != nil {
        return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
    }

    switch m.currentView {
    case mainView:
        return m.renderMainView()
    case detailView:
        return m.renderDetailView()
    case helpView:
        return m.renderHelpView()
    case statsView:
        return m.renderStatsView()
    case inputView:
        return m.input.View()
    case filterView:
        return m.renderFilterView()
    case confirmView:
        return m.renderConfirmView()
    default:
        return m.renderMainView()
    }
}

// renderMainView renders the main task list view
func (m Model) renderMainView() string {
    // Header with task count and filters
    header := headerStyle.Render("Task Manager")
    taskCount := infoStyle.Render(fmt.Sprintf(" | Tasks: %d", len(m.tasks)))
    header += taskCount

    // Active filters
    var activeFilters []string
    if m.filter.Status != nil {
        activeFilters = append(activeFilters, fmt.Sprintf("Status: %s", m.filter.Status.String()))
    }
    if m.filter.Priority != nil {
        activeFilters = append(activeFilters, fmt.Sprintf("Priority: %s", m.filter.Priority.String()))
    }
    if m.filter.Search != "" {
        activeFilters = append(activeFilters, fmt.Sprintf("Search: '%s'", m.filter.Search))
    }
    if m.filter.Tag != "" {
        activeFilters = append(activeFilters, fmt.Sprintf("Tag: %s", m.filter.Tag))
    }

    if len(activeFilters) > 0 {
        header += "\n" + filterStyle.Render("Filters: "+strings.Join(activeFilters, " | "))
    }

    // Status message
    statusBar := ""
    if m.statusMsg != "" {
        statusBar = lipgloss.NewStyle().
            Foreground(lipgloss.Color("39")).
            Render(m.statusMsg)
    }

    // Assemble view
    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        m.list.View(),
        statusBar,
        m.help.View(m.keys),
    )
}

// renderDetailView renders a detailed view of a single task
func (m Model) renderDetailView() string {
    task := m.list.SelectedTask()
    if task == nil {
        return "No task selected"
    }

    var details strings.Builder
    
    details.WriteString(titleStyle.Render(task.Title) + "\n\n")
    details.WriteString(fmt.Sprintf("Status: %s\n", task.Status.String()))
    details.WriteString(fmt.Sprintf("Priority: %s\n", task.Priority.String()))
    
    if task.DueDate != nil {
        details.WriteString(fmt.Sprintf("Due Date: %s\n", task.DueDate.Format("2006-01-02 15:04")))
    }
    
    if len(task.Tags) > 0 {
        details.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(task.Tags, ", ")))
    }
    
    details.WriteString(fmt.Sprintf("\nCreated: %s\n", task.CreatedAt.Format("2006-01-02 15:04")))
    details.WriteString(fmt.Sprintf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04")))
    
    if task.Description != "" {
        details.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.Description))
    }

    return borderStyle.Padding(1).Width(m.width-4).Render(details.String())
}

// renderHelpView renders the help screen
func (m Model) renderHelpView() string {
    helpText := `Keyboard Shortcuts
==================

Task Management
  n          - Create new task
  e          - Edit selected task
  d          - Delete selected task
  p          - Cycle priority of selected task
  t          - Cycle status of selected task
  enter      - View task details

Navigation
  up/down    - Navigate through tasks
  j/k        - Alternative navigation keys
  home/end   - Go to first/last task
  pgup/pgdn  - Page up/down

Filters and Search
  f          - Show filter options
  /          - Search tasks
  x          - Clear all filters
  esc        - Go back to main view

Other
  s          - View statistics
  ?          - Show this help
  q/ctrl+c   - Quit application

Press esc to return to task list`

    return borderStyle.Padding(1).Width(m.width-4).Render(helpText)
}

// renderStatsView renders the statistics screen
func (m Model) renderStatsView() string {
    var stats strings.Builder
    
    stats.WriteString(titleStyle.Render("Task Statistics") + "\n\n")
    
    if len(m.stats) == 0 {
        stats.WriteString("Loading statistics...")
    } else {
        stats.WriteString(fmt.Sprintf("Total Tasks:      %d\n", m.stats["total"]))
        stats.WriteString(fmt.Sprintf("Todo:             %d\n", m.stats["todo"]))
        stats.WriteString(fmt.Sprintf("In Progress:      %d\n", m.stats["inProgress"]))
        stats.WriteString(fmt.Sprintf("Done:             %d\n", m.stats["done"]))
        stats.WriteString(fmt.Sprintf("Archived:         %d\n", m.stats["archived"]))
        stats.WriteString(fmt.Sprintf("Overdue:          %d\n", m.stats["overdue"]))
        stats.WriteString(fmt.Sprintf("\nPriority Breakdown:\n"))
        stats.WriteString(fmt.Sprintf("  Urgent:  %d\n", m.stats["urgent"]))
        stats.WriteString(fmt.Sprintf("  High:    %d\n", m.stats["high"]))
        stats.WriteString(fmt.Sprintf("  Medium:  %d\n", m.stats["medium"]))
        stats.WriteString(fmt.Sprintf("  Low:     %d\n", m.stats["low"]))
    }
    
    stats.WriteString("\n\nPress esc to return")

    return borderStyle.Padding(1).Width(m.width-4).Render(stats.String())
}

// renderFilterView renders the filter selection screen
func (m Model) renderFilterView() string {
    return borderStyle.Padding(1).Width(m.width-4).Render(
        "Filter Tasks\n\n" +
        "Press 1-4 to filter by status:\n" +
        "  1 - Todo\n" +
        "  2 - In Progress\n" +
        "  3 - Done\n" +
        "  4 - Archived\n\n" +
        "Press a-d to filter by priority:\n" +
        "  a - Low\n" +
        "  b - Medium\n" +
        "  c - High\n" +
        "  d - Urgent\n\n" +
        "Press esc to return",
    )
}

// renderConfirmView renders a confirmation dialog
func (m Model) renderConfirmView() string {
    content := fmt.Sprintf("%s\n\nPress 'y' to confirm, 'n' or esc to cancel", m.statusMsg)
    return borderStyle.Padding(2).Width(60).Render(content)
}

// handleInputUpdate handles keyboard input when in input view
func (m Model) handleInputUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)

    if m.input.Done() {
        if !m.input.Canceled() {
            values := m.input.Values()
            return m, m.processInputValues(values)
        }
        m.currentView = mainView
        m.statusMsg = ""
        return m, nil
    }

    return m, cmd
}

// handleConfirmUpdate handles keyboard input when in confirmation view
func (m Model) handleConfirmUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y", "Y":
        task := m.list.SelectedTask()
        if task != nil {
            return m, m.deleteTaskCmd(task.ID)
        }
    case "n", "N", "esc":
        m.currentView = mainView
        m.statusMsg = ""
    }
    return m, nil
}

// showTaskForm opens the task creation/editing form
func (m *Model) showTaskForm(title, taskTitle, description, tags string) (tea.Model, tea.Cmd) {
    m.currentView = inputView
    
    fields := []string{"Title", "Description", "Tags (comma-separated)", "Due Date (YYYY-MM-DD)", "Priority (low/medium/high/urgent)"}
    m.input = components.NewInputModel(title, fields)
    m.input.SetSize(m.width, m.height)
    
    if taskTitle != "" {
        m.input.SetValues([]string{taskTitle, description, tags, "", ""})
    }
    
    return m, m.input.Init()
}

// showInputPrompt opens a simple single-field input form
func (m *Model) showInputPrompt(title, placeholder string) (tea.Model, tea.Cmd) {
    m.currentView = inputView
    m.input = components.NewInputModel(title, []string{placeholder})
    m.input.SetSize(m.width, m.height)
    return m, m.input.Init()
}

// processInputValues handles the submitted form values
func (m *Model) processInputValues(values []string) tea.Cmd {
    title := m.input.String()
    
    switch {
    case strings.Contains(title, "New Task"):
        return m.createTaskCmd(values)
        
    case strings.Contains(title, "Edit Task"):
        task := m.list.SelectedTask()
        if task != nil {
            return m.updateTaskCmd(task.ID, values)
        }
        
    case strings.Contains(title, "Search"):
        if len(values) > 0 {
            m.filter.Search = values[0]
            return m.loadTasksCmd()
        }
    }
    
    return nil
}

// Command functions
func (m *Model) loadTasksCmd() tea.Cmd {
    return func() tea.Msg {
        tasks, err := m.db.ListTasks(m.filter)
        if err != nil {
            return errorMsg{err}
        }
        return tasksLoadedMsg{tasks}
    }
}

func (m *Model) createTaskCmd(values []string) tea.Cmd {
    return func() tea.Msg {
        title := values[0]
        description := values[1]
        tags := parseTags(values[2])
        dueDate, _ := parseDueDate(values[3])
        priority := models.ParsePriority(values[4])

        task := &models.Task{
            Title:       title,
            Description: description,
            Priority:    priority,
            Status:      models.StatusTodo,
            DueDate:     dueDate,
            Tags:        tags,
        }

        if err := m.db.CreateTask(task); err != nil {
            return errorMsg{err}
        }
        return taskCreatedMsg{task}
    }
}

func (m *Model) updateTaskCmd(id int64, values []string) tea.Cmd {
    return func() tea.Msg {
        task, err := m.db.GetTask(id)
        if err != nil {
            return errorMsg{err}
        }

        task.Title = values[0]
        task.Description = values[1]
        task.Tags = parseTags(values[2])
        
        if dueDate, err := parseDueDate(values[3]); err == nil {
            task.DueDate = dueDate
        }
        
        task.Priority = models.ParsePriority(values[4])

        if err := m.db.UpdateTask(task); err != nil {
            return errorMsg{err}
        }
        return taskUpdatedMsg{task}
    }
}

func (m *Model) deleteTaskCmd(id int64) tea.Cmd {
    return func() tea.Msg {
        if err := m.db.DeleteTask(id); err != nil {
            return errorMsg{err}
        }
        return taskDeletedMsg{id}
    }
}

func (m *Model) loadStatsCmd() tea.Cmd {
    return func() tea.Msg {
        stats, err := m.db.GetStatistics()
        if err != nil {
            return errorMsg{err}
        }
        return statsLoadedMsg{stats}
    }
}

func (m *Model) cycleTaskPriority(id int64) tea.Cmd {
    return func() tea.Msg {
        task, err := m.db.GetTask(id)
        if err != nil {
            return errorMsg{err}
        }

        // Cycle through priorities
        task.Priority = (task.Priority + 1) % 4

        if err := m.db.UpdateTask(task); err != nil {
            return errorMsg{err}
        }
        return taskUpdatedMsg{task}
    }
}

func (m *Model) cycleTaskStatus(id int64) tea.Cmd {
    return func() tea.Msg {
        task, err := m.db.GetTask(id)
        if err != nil {
            return errorMsg{err}
        }

        // Cycle through statuses
        task.Status = (task.Status + 1) % 4

        if err := m.db.UpdateTask(task); err != nil {
            return errorMsg{err}
        }
        return taskUpdatedMsg{task}
    }
}