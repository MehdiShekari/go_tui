package ui

import (
    "strings"
    "time"

    "github.com/charmbracelet/lipgloss"
    "github.com/MehdiShekari/go_tui/models"
)

// Style definitions for the UI
var (
    titleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("205")).
            MarginBottom(1)

    headerStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("205")).
            Padding(0, 1)

    infoStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240"))

    filterStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("226")).
            Italic(true)

    borderStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("63"))

    errorStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true)
)

// custom message types for application events
type taskCreatedMsg struct {
    task *models.Task
}

type taskUpdatedMsg struct {
    task *models.Task
}

type taskDeletedMsg struct {
    id int64
}

type tasksLoadedMsg struct {
    tasks []models.Task
}

type statsLoadedMsg struct {
    stats map[string]int
}

type errorMsg struct {
    err error
}

// parseDueDate parses a date string in various formats
func parseDueDate(dateStr string) (*time.Time, error) {
    if dateStr == "" {
        return nil, nil
    }

    formats := []string{
        "2006-01-02",
        "01/02/2006",
        "2006-01-02 15:04",
        "01/02/2006 15:04",
        "Jan 2 2006",
        "January 2 2006",
    }

    for _, format := range formats {
        if t, err := time.Parse(format, dateStr); err == nil {
            return &t, nil
        }
    }

    return nil, nil
}

// parseTags splits a comma-separated string into a slice of tags
func parseTags(tagStr string) []string {
    if tagStr == "" {
        return nil
    }

    tags := strings.Split(tagStr, ",")
    result := make([]string, 0, len(tags))
    
    for _, tag := range tags {
        tag = strings.TrimSpace(tag)
        if tag != "" {
            result = append(result, tag)
        }
    }

    return result
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    if maxLen <= 3 {
        return s[:maxLen]
    }
    return s[:maxLen-3] + "..."
}

// priorityToColor returns a lipgloss color for a priority level
func priorityToColor(p models.Priority) lipgloss.Color {
    switch p {
    case models.PriorityLow:
        return lipgloss.Color("#00ff00")
    case models.PriorityMedium:
        return lipgloss.Color("#ffff00")
    case models.PriorityHigh:
        return lipgloss.Color("#ff8800")
    case models.PriorityUrgent:
        return lipgloss.Color("#ff0000")
    default:
        return lipgloss.Color("#ffffff")
    }
}

// statusToColor returns a lipgloss color for a status level
func statusToColor(s models.Status) lipgloss.Color {
    switch s {
    case models.StatusTodo:
        return lipgloss.Color("#aaaaaa")
    case models.StatusInProgress:
        return lipgloss.Color("#0088ff")
    case models.StatusDone:
        return lipgloss.Color("#00ff00")
    case models.StatusArchived:
        return lipgloss.Color("#666666")
    default:
        return lipgloss.Color("#ffffff")
    }
}