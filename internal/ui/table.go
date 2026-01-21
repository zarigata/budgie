package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column represents a table column
type Column struct {
	Title string
	Width int
}

// Row represents a table row
type Row []string

// Table is a container table view component
type Table struct {
	columns  []Column
	rows     []Row
	cursor   int
	offset   int
	height   int
	width    int
	focused  bool
}

// NewTable creates a new table
func NewTable(columns []Column) Table {
	return Table{
		columns: columns,
		rows:    make([]Row, 0),
		cursor:  0,
		height:  10, // Default visible rows
	}
}

// SetRows sets the table rows
func (t *Table) SetRows(rows []Row) {
	t.rows = rows
	if t.cursor >= len(rows) {
		t.cursor = len(rows) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row Row) {
	t.rows = append(t.rows, row)
}

// ClearRows clears all rows
func (t *Table) ClearRows() {
	t.rows = make([]Row, 0)
	t.cursor = 0
	t.offset = 0
}

// Init initializes the table
func (t Table) Init() tea.Cmd {
	return nil
}

// Update handles table input
func (t Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	if !t.focused {
		return t, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
				if t.cursor < t.offset {
					t.offset = t.cursor
				}
			}
		case "down", "j":
			if t.cursor < len(t.rows)-1 {
				t.cursor++
				if t.cursor >= t.offset+t.height {
					t.offset = t.cursor - t.height + 1
				}
			}
		case "home", "g":
			t.cursor = 0
			t.offset = 0
		case "end", "G":
			t.cursor = len(t.rows) - 1
			if t.cursor >= t.height {
				t.offset = t.cursor - t.height + 1
			}
		case "pgup":
			t.cursor -= t.height
			if t.cursor < 0 {
				t.cursor = 0
			}
			t.offset = t.cursor
		case "pgdown":
			t.cursor += t.height
			if t.cursor >= len(t.rows) {
				t.cursor = len(t.rows) - 1
			}
			if t.cursor >= t.offset+t.height {
				t.offset = t.cursor - t.height + 1
			}
		}
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height - 5 // Leave room for header and footer
		if t.height < 1 {
			t.height = 1
		}
	}

	return t, nil
}

// View renders the table
func (t Table) View() string {
	var b strings.Builder

	// Header
	var headerCells []string
	for _, col := range t.columns {
		cell := TableHeaderStyle.Width(col.Width).Render(col.Title)
		headerCells = append(headerCells, cell)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Separator
	totalWidth := 0
	for _, col := range t.columns {
		totalWidth += col.Width
	}
	b.WriteString(strings.Repeat("─", totalWidth))
	b.WriteString("\n")

	// Rows
	end := t.offset + t.height
	if end > len(t.rows) {
		end = len(t.rows)
	}

	if len(t.rows) == 0 {
		b.WriteString(SubtitleStyle.Render("  No data"))
		b.WriteString("\n")
	} else {
		for i := t.offset; i < end; i++ {
			row := t.rows[i]
			var cells []string

			for j, col := range t.columns {
				var cellContent string
				if j < len(row) {
					cellContent = row[j]
				}

				style := TableCellStyle.Width(col.Width)
				if i == t.cursor && t.focused {
					style = style.Foreground(Primary).Bold(true)
				}
				cells = append(cells, style.Render(cellContent))
			}

			if i == t.cursor && t.focused {
				b.WriteString("> ")
			} else {
				b.WriteString("  ")
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
			b.WriteString("\n")
		}
	}

	// Footer with scroll info
	if len(t.rows) > t.height {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d", t.offset+1, end, len(t.rows))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(scrollInfo))
	}

	return b.String()
}

// Focus sets the table focus
func (t *Table) Focus() {
	t.focused = true
}

// Blur removes the table focus
func (t *Table) Blur() {
	t.focused = false
}

// Focused returns whether the table is focused
func (t Table) Focused() bool {
	return t.focused
}

// Cursor returns the current cursor position
func (t Table) Cursor() int {
	return t.cursor
}

// SelectedRow returns the currently selected row
func (t Table) SelectedRow() Row {
	if t.cursor >= 0 && t.cursor < len(t.rows) {
		return t.rows[t.cursor]
	}
	return nil
}

// SetHeight sets the visible height
func (t *Table) SetHeight(h int) {
	t.height = h
}

// Rows returns all rows
func (t Table) Rows() []Row {
	return t.rows
}

// RowCount returns the number of rows
func (t Table) RowCount() int {
	return len(t.rows)
}
