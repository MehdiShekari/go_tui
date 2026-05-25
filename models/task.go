package models

import (
    "time"
)

// Priority represents the importance level of a task
type Priority int

const (
    PriorityLow Priority = iota
    PriorityMedium
    PriorityHigh
    PriorityUrgent
)

// String returns the string representation of Priority
func (p Priority) String() string {
    switch p {
    case PriorityLow:
        return "Low"
    case PriorityMedium:
        return "Medium"
    case PriorityHigh:
        return "High"
    case PriorityUrgent:
        return "Urgent"
    default:
        return "Unknown"
    }
}

// Color returns the hex color code for the priority level
func (p Priority) Color() string {
    switch p {
    case PriorityLow:
        return "#00ff00"
    case PriorityMedium:
        return "#ffff00"
    case PriorityHigh:
        return "#ff8800"
    case PriorityUrgent:
        return "#ff0000"
    default:
        return "#ffffff"
    }
}

// ParsePriority converts a string to Priority type
func ParsePriority(s string) Priority {
    switch s {
    case "low", "l":
        return PriorityLow
    case "medium", "m":
        return PriorityMedium
    case "high", "h":
        return PriorityHigh
    case "urgent", "u":
        return PriorityUrgent
    default:
        return PriorityMedium
    }
}

// Status represents the current state of a task
type Status int

const (
    StatusTodo Status = iota
    StatusInProgress
    StatusDone
    StatusArchived
)

// String returns the string representation of Status
func (s Status) String() string {
    switch s {
    case StatusTodo:
        return "Todo"
    case StatusInProgress:
        return "In Progress"
    case StatusDone:
        return "Done"
    case StatusArchived:
        return "Archived"
    default:
        return "Unknown"
    }
}

// Icon returns a text icon for the status
func (s Status) Icon() string {
    switch s {
    case StatusTodo:
        return "[ ]"
    case StatusInProgress:
        return "[~]"
    case StatusDone:
        return "[x]"
    case StatusArchived:
        return "[-]"
    default:
        return "[?]"
    }
}

// ParseStatus converts a string to Status type
func ParseStatus(s string) Status {
    switch s {
    case "todo", "td":
        return StatusTodo
    case "inprogress", "in-progress", "ip", "progress":
        return StatusInProgress
    case "done", "d":
        return StatusDone
    case "archived", "a":
        return StatusArchived
    default:
        return StatusTodo
    }
}

// Task represents the core task structure
type Task struct {
    ID          int64      `json:"id"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Priority    Priority   `json:"priority"`
    Status      Status     `json:"status"`
    DueDate     *time.Time `json:"due_date,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    Tags        []string   `json:"tags"`
    ParentID    *int64     `json:"parent_id,omitempty"`
}

// TaskFilter defines filtering criteria for task queries
type TaskFilter struct {
    Status   *Status
    Priority *Priority
    Search   string
    Tag      string
}

// IsOverdue checks if a task is past its due date
func (t *Task) IsOverdue() bool {
    if t.DueDate == nil || t.Status == StatusDone || t.Status == StatusArchived {
        return false
    }
    return time.Now().After(*t.DueDate)
}

// IsDueToday checks if a task is due today
func (t *Task) IsDueToday() bool {
    if t.DueDate == nil {
        return false
    }
    now := time.Now()
    dueDate := *t.DueDate
    return dueDate.Year() == now.Year() &&
        dueDate.Month() == now.Month() &&
        dueDate.Day() == now.Day()
}