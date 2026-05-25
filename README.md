# Go Task Manager TUI

A highly responsive, terminal-based productivity interface built with Go. Powered by the **Charmbracelet Bubble Tea** framework, this application utilizes an architecture that features multi-field form traversal, strict validation routines, live query filters, and persistent local SQLite storage.

---

## Key Features

* **Keyboard-Centric Data Entry:** Easily switch and update fields via `Tab`, `Shift+Tab`, or arrow loops.
* **Live Input Validation:** Real-time feedback rendered instantly below input targets for constraints (e.g., minimum character checking, date formatting rules like `YYYY-MM-DD`, and strict enum priority values).
* **Comprehensive Task States:** Seamlessly cycle priority levels (`low`, `medium`, `high`, `urgent`) and task states directly from the master list component view.
* **Interactive Search & Filter:** Dynamically query and slice your tasks on-screen. Pressing `Esc` at any point quickly clears active state queries to return safely to the overarching task dashboard menu.
* **Persistent SQLite Backend:** Automatic database generation and tracking to prevent context or record data loss across application lifecycles.

---

## Navigation & Controls Map

### Dashboard View
| Keystroke | Action Description |
| :--- | :--- |
| `n` | Initialize new task creation form |
| `e` | Modify selected task properties |
| `d` | Permanent deletion of chosen record |
| `Space` | Cycle through task progress states (`Todo` $\rightarrow$ `In Progress` $\rightarrow$ `Done`) |
| `p` | Cycle priority variant tiers (`Low` $\rightarrow$ `Medium` $\rightarrow$ `High` $\rightarrow$ `Urgent`) |
| `/` | Invoke search query input window |
| `x` | Wipe active selection criteria/filters |
| `Enter` | Open full detail viewer overlay for highlighted task |
| `Esc` | Clear matching keywords filter and reset row focus position |
| `q` | Safe termination of application thread |

### Form / Input View
| Keystroke | Action Description |
| :--- | :--- |
| `Tab` / `Down` | Shift focus down to the next input line field |
| `Shift+Tab` / `Up` | Revert focus upward to previous input line field |
| `Enter` | Move down an element, or submit form payload if on the final field |
| `Esc` | Discard structural edits and return immediately to main menu |

---

## Installation & Getting Started

### Prerequisites
* Go 1.18 or higher installed on your target execution platform.
* SQLite system binaries or localized environment runtime compilation tooling.

### Installation
Clone this repository directly from source control and traverse into the workspace folder root:
```bash
git clone https://github.com/MehdiShekari/go_tui.git
cd go_tui
```
