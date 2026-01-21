package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a menu item
type MenuItem struct {
	Label       string
	Description string
	Icon        string
	Action      func() tea.Cmd
}

// Menu is an interactive menu component
type Menu struct {
	title    string
	items    []MenuItem
	cursor   int
	selected int
	width    int
	height   int
}

// NewMenu creates a new menu
func NewMenu(title string, items []MenuItem) Menu {
	return Menu{
		title:    title,
		items:    items,
		cursor:   0,
		selected: -1,
	}
}

// Init initializes the menu
func (m Menu) Init() tea.Cmd {
	return nil
}

// Update handles menu input
func (m Menu) Update(msg tea.Msg) (Menu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			if m.items[m.cursor].Action != nil {
				return m, m.items[m.cursor].Action()
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.items) - 1
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the menu
func (m Menu) View() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Items
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		icon := item.Icon
		if icon == "" {
			icon = "•"
		}

		var line string
		if i == m.cursor {
			line = SelectedMenuItemStyle.Render(fmt.Sprintf("%s%s %s", cursor, icon, item.Label))
		} else {
			line = MenuItemStyle.Render(fmt.Sprintf("%s%s %s", cursor, icon, item.Label))
		}

		b.WriteString(line)

		// Show description for selected item
		if i == m.cursor && item.Description != "" {
			b.WriteString("\n")
			b.WriteString(SubtitleStyle.Render("    " + item.Description))
		}

		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: navigate • enter: select • q: quit"))

	return b.String()
}

// Selected returns the selected item index
func (m Menu) Selected() int {
	return m.selected
}

// Cursor returns the current cursor position
func (m Menu) Cursor() int {
	return m.cursor
}

// SetCursor sets the cursor position
func (m *Menu) SetCursor(pos int) {
	if pos >= 0 && pos < len(m.items) {
		m.cursor = pos
	}
}

// Items returns the menu items
func (m Menu) Items() []MenuItem {
	return m.items
}

// SelectionBox renders a selection box with options
type SelectionBox struct {
	title    string
	options  []string
	selected int
	style    lipgloss.Style
}

// NewSelectionBox creates a new selection box
func NewSelectionBox(title string, options []string) SelectionBox {
	return SelectionBox{
		title:    title,
		options:  options,
		selected: 0,
		style:    BoxStyle,
	}
}

// Update handles selection box input
func (s SelectionBox) Update(msg tea.Msg) (SelectionBox, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.selected > 0 {
				s.selected--
			}
		case "down", "j":
			if s.selected < len(s.options)-1 {
				s.selected++
			}
		}
	}
	return s, nil
}

// View renders the selection box
func (s SelectionBox) View() string {
	var b strings.Builder

	b.WriteString(s.title)
	b.WriteString("\n\n")

	for i, opt := range s.options {
		if i == s.selected {
			b.WriteString(SelectedMenuItemStyle.Render(fmt.Sprintf(" > %s", opt)))
		} else {
			b.WriteString(MenuItemStyle.Render(fmt.Sprintf("   %s", opt)))
		}
		b.WriteString("\n")
	}

	return s.style.Render(b.String())
}

// Selected returns the selected option index
func (s SelectionBox) Selected() int {
	return s.selected
}

// SelectedOption returns the selected option string
func (s SelectionBox) SelectedOption() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected]
	}
	return ""
}
