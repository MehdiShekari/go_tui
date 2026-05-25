package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	database "github.com/MehdiShekari/go_tui/db"
	"github.com/MehdiShekari/go_tui/models"
	"github.com/MehdiShekari/go_tui/ui/components"
)

// View represents the different screen states of the application
type View int

const (
	mainView View = iota
	detailView
	helpView
	statsView
	inputView
	filterView
	confirmView
)

type keyMap struct {
	Quit        key.Binding
	Help        key.Binding
	Back        key.Binding
	New         key.Binding
	Edit        key.Binding
	Delete      key.Binding
	ToggleState key.Binding
	CyclePri    key.Binding
	Search      key.Binding
	ClearFilter key.Binding
}

var keys = keyMap{
	Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Back:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	New:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new task")),
	Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit task")),
	Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete task")),
	ToggleState: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle status")),
	CyclePri:    key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "cycle priority")),
	Search:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	ClearFilter: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "clear filters")),
}

// Model represents the main application model state
type Model struct {
	db          *database.Database
	list        components.TaskListModel
	input       components.InputModel
	help        help.Model
	keys        keyMap
	currentView View
	tasks       []models.Task
	filter      models.TaskFilter
	width       int
	height      int
	ready       bool
	err         error
	stats       map[string]int
	statusMsg   string
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

func (m Model) Init() tea.Cmd {
	return m.loadTasksCmd()
}

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

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.list.SetTasks(m.tasks)
		return m, nil

	case taskCreatedMsg, taskUpdatedMsg, taskDeletedMsg:
		m.statusMsg = "Operation completed successfully!"
		return m, m.loadTasksCmd()

	case statsLoadedMsg:
		m.stats = msg.stats
		m.currentView = statsView
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case tea.KeyMsg:
		if m.currentView == inputView {
			return m.handleInputUpdate(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.currentView = helpView
			return m, nil
		case key.Matches(msg, m.keys.Back):
			// ESC completely resets search filters and returns to full menu list
			m.currentView = mainView
			m.filter = models.TaskFilter{}
			m.statusMsg = ""
			m.list.ResetCursor()
			return m, m.loadTasksCmd()
		case msg.String() == "enter":
			// If on main view and press enter, view selected task detail
			if m.currentView == mainView {
				task := m.list.SelectedTask()
				if task != nil {
					m.currentView = detailView
					return m, nil
				}
			}
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
				return m, m.deleteTaskCmd(task.ID)
			}
		case key.Matches(msg, m.keys.ToggleState):
			task := m.list.SelectedTask()
			if task != nil {
				return m, m.cycleTaskStatus(task.ID)
			}
		case key.Matches(msg, m.keys.CyclePri):
			task := m.list.SelectedTask()
			if task != nil {
				return m, m.cycleTaskPriority(task.ID)
			}
		case key.Matches(msg, m.keys.Search):
			return m.showInputPrompt("Search Tasks", "Enter keywords...")
		case key.Matches(msg, m.keys.ClearFilter):
			m.filter = models.TaskFilter{}
			m.statusMsg = "Filters cleared."
			m.list.ResetCursor()
			return m, m.loadTasksCmd()
		}
	}

	// Only process list controls if we are looking at the list menu
	if m.currentView == mainView {
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleInputUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	if m.input.Done() {
		if m.input.Canceled() {
			m.currentView = mainView
			m.statusMsg = "Form operation discarded."
			return m, m.loadTasksCmd()
		}

		values := m.input.GetValues()
		m.input.ClearErrors()
		hasValidationFailures := false

		if strings.Contains(m.input.Title(), "Task") && !strings.Contains(m.input.Title(), "Search") {
			titleVal := strings.TrimSpace(values[0])
			if len(titleVal) < 3 {
				m.input.SetFieldError(0, "Constraint failure: Title must contain at least 3 characters.")
				hasValidationFailures = true
			} else if len(titleVal) > 50 {
				m.input.SetFieldError(0, "Constraint failure: Title must be 50 characters or less.")
				hasValidationFailures = true
			}

			if len(values) > 3 && strings.TrimSpace(values[3]) != "" {
				_, err := time.Parse("2006-01-02", strings.TrimSpace(values[3]))
				if err != nil {
					m.input.SetFieldError(3, "Structure mistake: Date format must conform to YYYY-MM-DD.")
					hasValidationFailures = true
				}
			}

			if len(values) > 4 && strings.TrimSpace(values[4]) != "" {
				priorityInput := strings.ToLower(strings.TrimSpace(values[4]))
				if priorityInput != "low" && priorityInput != "medium" && priorityInput != "high" && priorityInput != "urgent" {
					m.input.SetFieldError(4, "Type variant warning: Priority must equal low, medium, high, or urgent.")
					hasValidationFailures = true
				}
			}
		}

		if hasValidationFailures {
			m.input.ResetDoneStatus()
			m.currentView = inputView
			return m, nil
		}

		m.currentView = mainView
		return m, m.processInputValues(values)
	}

	return m, cmd
}

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
	default:
		return m.renderMainView()
	}
}

func (m Model) renderMainView() string {
	header := headerStyle.Render("Task Manager")
	taskCount := infoStyle.Render(fmt.Sprintf(" | Tasks: %d", len(m.tasks)))
	header += taskCount

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
	if len(activeFilters) > 0 {
		header += "\n" + filterStyle.Render("Filters: "+strings.Join(activeFilters, " | "))
	}

	statusBar := ""
	if m.statusMsg != "" {
		statusBar = "\n" + filterStyle.Render(m.statusMsg)
	}

	return borderStyle.Padding(1).Width(m.width - 4).Render(header + "\n\n" + m.list.View() + statusBar)
}

func (m Model) renderDetailView() string {
	task := m.list.SelectedTask()
	if task == nil {
		return "No task selected. Press esc to go back."
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Task Detail: %s", task.Title)) + "\n\n")
	b.WriteString(fmt.Sprintf("Description: %s\n", task.Description))
	b.WriteString(fmt.Sprintf("Status:      %s\n", task.Status.String()))
	b.WriteString(fmt.Sprintf("Priority:    %s\n", task.Priority.String()))
	
	if task.DueDate != nil {
		b.WriteString(fmt.Sprintf("Due Date:    %s\n", task.DueDate.Format("2006-01-02")))
	}
	if len(task.Tags) > 0 {
		b.WriteString(fmt.Sprintf("Tags:        %s\n", strings.Join(task.Tags, ", ")))
	}

	b.WriteString("\n\n" + infoStyle.Render("Press esc to return to full menu list"))
	return borderStyle.Padding(1).Width(m.width - 4).Render(b.String())
}

func (m Model) renderHelpView() string {
	helpText := `Help & Controls:
n         - New task form
e         - Edit selected task
d         - Delete selected task
space     - Cycle status of selected task
p         - Cycle priority of selected task
/         - Search tasks
x         - Clear all filters
esc       - Go back to main view / clear search filters
q         - Quit application`
	return borderStyle.Padding(1).Width(m.width - 4).Render(helpText)
}

func (m Model) renderStatsView() string {
	var stats strings.Builder
	stats.WriteString(titleStyle.Render("Task Statistics") + "\n\n")
	if len(m.stats) == 0 {
		stats.WriteString("Loading statistics...")
	} else {
		stats.WriteString(fmt.Sprintf("Total Tasks: %d\n", m.stats["total"]))
	}
	stats.WriteString("\n\nPress esc to return")
	return borderStyle.Padding(1).Width(m.width - 4).Render(stats.String())
}

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

func (m *Model) showInputPrompt(title, placeholder string) (tea.Model, tea.Cmd) {
	m.currentView = inputView
	m.input = components.NewInputModel(title, []string{placeholder})
	m.input.SetSize(m.width, m.height)
	return m, m.input.Init()
}

func (m *Model) processInputValues(values []string) tea.Cmd {
	formTitle := m.input.Title()
	switch {
	case strings.Contains(formTitle, "New Task"):
		return m.createTaskCmd(values)
	case strings.Contains(formTitle, "Edit Task"):
		task := m.list.SelectedTask()
		if task != nil {
			return m.updateTaskCmd(task.ID, values)
		}
	case strings.Contains(formTitle, "Search"):
		if len(values) > 0 {
			m.filter.Search = strings.TrimSpace(values[0])
			m.list.ResetCursor() // Safely bring cursor back to top item
			return m.loadTasksCmd()
		}
	}
	return nil
}

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
		task := &models.Task{
			Title:       values[0],
			Description: values[1],
			Status:      models.StatusTodo,
		}
		if len(values) > 2 {
			task.Tags = parseTags(values[2])
		}
		if len(values) > 3 && values[3] != "" {
			if dueDate, err := parseDueDate(values[3]); err == nil {
				task.DueDate = dueDate
			}
		}
		if len(values) > 4 && values[4] != "" {
			task.Priority = models.ParsePriority(values[4])
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
		if len(values) > 2 {
			task.Tags = parseTags(values[2])
		}
		if len(values) > 3 && values[3] != "" {
			if dueDate, err := parseDueDate(values[3]); err == nil {
				task.DueDate = dueDate
			}
		}
		if len(values) > 4 && values[4] != "" {
			task.Priority = models.ParsePriority(values[4])
		}
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

func (m *Model) cycleTaskPriority(id int64) tea.Cmd {
	return func() tea.Msg {
		task, err := m.db.GetTask(id)
		if err != nil {
			return errorMsg{err}
		}
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
		task.Status = (task.Status + 1) % 4
		if err := m.db.UpdateTask(task); err != nil {
			return errorMsg{err}
		}
		return taskUpdatedMsg{task}
	}
}