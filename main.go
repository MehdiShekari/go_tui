package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/MehdiShekari/go_tui/db"
    "github.com/MehdiShekari/go_tui/ui"
)

func main() {
    // Determine database path from command line argument or use default
    dbPath := "tasks.db"
    if len(os.Args) > 1 {
        dbPath = os.Args[1]
    }

    // Initialize database connection
    db, err := database.NewDatabase(dbPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()

    // Create and run the TUI application
    app := ui.NewModel(db)
    program := tea.NewProgram(
        app,
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )

    if _, err := program.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
        os.Exit(1)
    }
}