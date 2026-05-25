package database

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "github.com/MehdiShekari/go_tui/models"
)

// Database wraps the SQLite connection and provides data access methods
type Database struct {
    db *sql.DB
}

// NewDatabase creates a new Database instance and initializes the schema
func NewDatabase(dbPath string) (*Database, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Enable WAL mode for better concurrent access
    if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
        return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
    }

    // Enable foreign key support
    if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
        return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    database := &Database{db: db}
    if err := database.createSchema(); err != nil {
        return nil, fmt.Errorf("failed to create schema: %w", err)
    }

    return database, nil
}

// createSchema initializes the database tables and indexes
func (d *Database) createSchema() error {
    schema := []string{
        `CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            description TEXT DEFAULT '',
            priority INTEGER DEFAULT 1,
            status INTEGER DEFAULT 0,
            due_date DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            parent_id INTEGER,
            FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL
        )`,
        
        `CREATE TABLE IF NOT EXISTS tags (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE COLLATE NOCASE
        )`,
        
        `CREATE TABLE IF NOT EXISTS task_tags (
            task_id INTEGER NOT NULL,
            tag_id INTEGER NOT NULL,
            PRIMARY KEY (task_id, tag_id),
            FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
            FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
        )`,
        
        `CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
        `CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)`,
        `CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date)`,
        `CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)`,
        `CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id)`,
        `CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_id)`,
    }

    for _, query := range schema {
        if _, err := d.db.Exec(query); err != nil {
            return fmt.Errorf("failed to execute schema query: %w", err)
        }
    }

    return nil
}

// CreateTask inserts a new task and its associated tags into the database
func (d *Database) CreateTask(task *models.Task) error {
    tx, err := d.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    now := time.Now()
    task.CreatedAt = now
    task.UpdatedAt = now

    result, err := tx.Exec(
        `INSERT INTO tasks (title, description, priority, status, due_date, parent_id)
         VALUES (?, ?, ?, ?, ?, ?)`,
        task.Title, task.Description, task.Priority, task.Status,
        task.DueDate, task.ParentID,
    )
    if err != nil {
        return fmt.Errorf("failed to insert task: %w", err)
    }

    taskID, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("failed to get last insert ID: %w", err)
    }
    task.ID = taskID

    // Insert tags
    if err := d.syncTags(tx, task.ID, task.Tags); err != nil {
        return fmt.Errorf("failed to sync tags: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// UpdateTask modifies an existing task and its tags
func (d *Database) UpdateTask(task *models.Task) error {
    tx, err := d.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    task.UpdatedAt = time.Now()

    result, err := tx.Exec(
        `UPDATE tasks 
         SET title = ?, description = ?, priority = ?, status = ?, 
             due_date = ?, parent_id = ?, updated_at = ?
         WHERE id = ?`,
        task.Title, task.Description, task.Priority, task.Status,
        task.DueDate, task.ParentID, task.UpdatedAt, task.ID,
    )
    if err != nil {
        return fmt.Errorf("failed to update task: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("task with ID %d not found", task.ID)
    }

    // Remove existing tags and insert new ones
    if _, err := tx.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID); err != nil {
        return fmt.Errorf("failed to delete existing tags: %w", err)
    }

    if err := d.syncTags(tx, task.ID, task.Tags); err != nil {
        return fmt.Errorf("failed to sync tags: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// DeleteTask removes a task and its associated tags from the database
func (d *Database) DeleteTask(id int64) error {
    result, err := d.db.Exec("DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        return fmt.Errorf("failed to delete task: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("task with ID %d not found", id)
    }

    return nil
}

// GetTask retrieves a single task by its ID, including its tags
func (d *Database) GetTask(id int64) (*models.Task, error) {
    task := &models.Task{}
    var dueDate sql.NullTime
    var parentID sql.NullInt64

    err := d.db.QueryRow(
        `SELECT id, title, description, priority, status, due_date, 
                created_at, updated_at, parent_id
         FROM tasks WHERE id = ?`, id,
    ).Scan(&task.ID, &task.Title, &task.Description, &task.Priority,
        &task.Status, &dueDate, &task.CreatedAt, &task.UpdatedAt, &parentID)

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("task with ID %d not found", id)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to scan task: %w", err)
    }

    if dueDate.Valid {
        task.DueDate = &dueDate.Time
    }
    if parentID.Valid {
        task.ParentID = &parentID.Int64
    }

    tags, err := d.getTaskTags(task.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get task tags: %w", err)
    }
    task.Tags = tags

    return task, nil
}

// ListTasks retrieves tasks based on the provided filter criteria
func (d *Database) ListTasks(filter models.TaskFilter) ([]models.Task, error) {
    query := `SELECT id, title, description, priority, status, due_date, 
              created_at, updated_at, parent_id FROM tasks WHERE 1=1`
    args := []interface{}{}

    if filter.Status != nil {
        query += " AND status = ?"
        args = append(args, int(*filter.Status))
    }

    if filter.Priority != nil {
        query += " AND priority = ?"
        args = append(args, int(*filter.Priority))
    }

    if filter.Search != "" {
        query += " AND (title LIKE ? OR description LIKE ?)"
        searchTerm := "%" + filter.Search + "%"
        args = append(args, searchTerm, searchTerm)
    }

    if filter.Tag != "" {
        query += ` AND id IN (
            SELECT task_id FROM task_tags tt 
            JOIN tags t ON tt.tag_id = t.id 
            WHERE t.name = ?
        )`
        args = append(args, filter.Tag)
    }

    query += " ORDER BY priority DESC, due_date ASC NULLS LAST, created_at DESC"

    rows, err := d.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query tasks: %w", err)
    }
    defer rows.Close()

    var tasks []models.Task
    for rows.Next() {
        var task models.Task
        var dueDate sql.NullTime
        var parentID sql.NullInt64

        err := rows.Scan(&task.ID, &task.Title, &task.Description,
            &task.Priority, &task.Status, &dueDate,
            &task.CreatedAt, &task.UpdatedAt, &parentID)
        if err != nil {
            return nil, fmt.Errorf("failed to scan task row: %w", err)
        }

        if dueDate.Valid {
            task.DueDate = &dueDate.Time
        }
        if parentID.Valid {
            task.ParentID = &parentID.Int64
        }

        tags, err := d.getTaskTags(task.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get tags for task %d: %w", task.ID, err)
        }
        task.Tags = tags

        tasks = append(tasks, task)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating task rows: %w", err)
    }

    return tasks, nil
}

// GetSubtasks retrieves all subtasks for a given parent task
func (d *Database) GetSubtasks(parentID int64) ([]models.Task, error) {
    rows, err := d.db.Query(
        `SELECT id, title, description, priority, status, due_date, 
                created_at, updated_at, parent_id
         FROM tasks WHERE parent_id = ?
         ORDER BY priority DESC, due_date ASC`,
        parentID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query subtasks: %w", err)
    }
    defer rows.Close()

    var tasks []models.Task
    for rows.Next() {
        var task models.Task
        var dueDate sql.NullTime
        var parentID sql.NullInt64

        err := rows.Scan(&task.ID, &task.Title, &task.Description,
            &task.Priority, &task.Status, &dueDate,
            &task.CreatedAt, &task.UpdatedAt, &parentID)
        if err != nil {
            return nil, fmt.Errorf("failed to scan subtask row: %w", err)
        }

        if dueDate.Valid {
            task.DueDate = &dueDate.Time
        }
        if parentID.Valid {
            task.ParentID = &parentID.Int64
        }

        tags, err := d.getTaskTags(task.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get tags for subtask %d: %w", task.ID, err)
        }
        task.Tags = tags

        tasks = append(tasks, task)
    }

    return tasks, nil
}

// GetAllTags retrieves all unique tag names from the database
func (d *Database) GetAllTags() ([]string, error) {
    rows, err := d.db.Query("SELECT name FROM tags ORDER BY name")
    if err != nil {
        return nil, fmt.Errorf("failed to query tags: %w", err)
    }
    defer rows.Close()

    var tags []string
    for rows.Next() {
        var tag string
        if err := rows.Scan(&tag); err != nil {
            return nil, fmt.Errorf("failed to scan tag: %w", err)
        }
        tags = append(tags, tag)
    }

    return tags, nil
}

// GetStatistics returns various statistics about the tasks
func (d *Database) GetStatistics() (map[string]int, error) {
    stats := make(map[string]int)

    queries := map[string]string{
        "total":      "SELECT COUNT(*) FROM tasks",
        "todo":       "SELECT COUNT(*) FROM tasks WHERE status = 0",
        "inProgress": "SELECT COUNT(*) FROM tasks WHERE status = 1",
        "done":       "SELECT COUNT(*) FROM tasks WHERE status = 2",
        "archived":   "SELECT COUNT(*) FROM tasks WHERE status = 3",
        "urgent":     "SELECT COUNT(*) FROM tasks WHERE priority = 3",
        "high":       "SELECT COUNT(*) FROM tasks WHERE priority = 2",
        "medium":     "SELECT COUNT(*) FROM tasks WHERE priority = 1",
        "low":        "SELECT COUNT(*) FROM tasks WHERE priority = 0",
    }

    for key, query := range queries {
        var count int
        err := d.db.QueryRow(query).Scan(&count)
        if err != nil {
            return nil, fmt.Errorf("failed to get %s count: %w", key, err)
        }
        stats[key] = count
    }

    // Count overdue tasks - Fixed: use temporary variable instead of taking address of map value
    var overdueCount int
    err := d.db.QueryRow(
        "SELECT COUNT(*) FROM tasks WHERE status IN (0, 1) AND due_date < datetime('now')",
    ).Scan(&overdueCount)
    if err != nil {
        return nil, fmt.Errorf("failed to get overdue count: %w", err)
    }
    stats["overdue"] = overdueCount

    return stats, nil
}

// syncTags ensures all tags exist and associates them with a task
func (d *Database) syncTags(tx *sql.Tx, taskID int64, tags []string) error {
    for _, tagName := range tags {
        if tagName == "" {
            continue
        }

        // Get or create tag
        var tagID int64
        err := tx.QueryRow(
            "SELECT id FROM tags WHERE name = ?", tagName,
        ).Scan(&tagID)

        if err == sql.ErrNoRows {
            result, err := tx.Exec(
                "INSERT INTO tags (name) VALUES (?)", tagName,
            )
            if err != nil {
                return fmt.Errorf("failed to insert tag '%s': %w", tagName, err)
            }
            tagID, err = result.LastInsertId()
            if err != nil {
                return fmt.Errorf("failed to get tag ID for '%s': %w", tagName, err)
            }
        } else if err != nil {
            return fmt.Errorf("failed to query tag '%s': %w", tagName, err)
        }

        // Create association
        _, err = tx.Exec(
            "INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)",
            taskID, tagID,
        )
        if err != nil {
            return fmt.Errorf("failed to associate tag '%s' with task: %w", tagName, err)
        }
    }

    return nil
}

// getTaskTags retrieves all tags associated with a task
func (d *Database) getTaskTags(taskID int64) ([]string, error) {
    rows, err := d.db.Query(
        `SELECT t.name FROM tags t
         JOIN task_tags tt ON t.id = tt.tag_id
         WHERE tt.task_id = ?
         ORDER BY t.name`,
        taskID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query task tags: %w", err)
    }
    defer rows.Close()

    var tags []string
    for rows.Next() {
        var tag string
        if err := rows.Scan(&tag); err != nil {
            return nil, fmt.Errorf("failed to scan tag: %w", err)
        }
        tags = append(tags, tag)
    }

    return tags, nil
}

// Close closes the database connection
func (d *Database) Close() error {
    return d.db.Close()
}