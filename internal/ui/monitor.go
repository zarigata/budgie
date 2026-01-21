package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ContainerStats represents container statistics
type ContainerStats struct {
	ID          string
	Name        string
	Status      string
	CPU         float64 // Percentage
	Memory      int64   // Bytes
	MemoryLimit int64   // Bytes
	NetRx       int64   // Bytes received
	NetTx       int64   // Bytes sent
	BlockRead   int64   // Bytes read
	BlockWrite  int64   // Bytes written
	PIDs        int
}

// TickMsg triggers a stats refresh
type TickMsg time.Time

// Monitor is a real-time monitoring dashboard
type Monitor struct {
	containers []ContainerStats
	table      Table
	width      int
	height     int
	interval   time.Duration
	ticker     *time.Ticker
}

// NewMonitor creates a new monitoring dashboard
func NewMonitor(interval time.Duration) Monitor {
	columns := []Column{
		{Title: "ID", Width: 14},
		{Title: "NAME", Width: 15},
		{Title: "STATUS", Width: 10},
		{Title: "CPU", Width: 8},
		{Title: "MEM", Width: 12},
		{Title: "NET I/O", Width: 16},
		{Title: "PIDS", Width: 6},
	}

	table := NewTable(columns)
	table.Focus()

	return Monitor{
		containers: make([]ContainerStats, 0),
		table:      table,
		interval:   interval,
	}
}

// Init initializes the monitor
func (m Monitor) Init() tea.Cmd {
	return tickCmd(m.interval)
}

// tickCmd creates a tick command
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles monitor input
func (m Monitor) Update(msg tea.Msg) (Monitor, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Force refresh
			return m, tickCmd(0)
		}
		// Pass to table
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case TickMsg:
		// Update stats and schedule next tick
		return m, tickCmd(m.interval)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(msg.Height - 8)
	}

	return m, nil
}

// UpdateStats updates the container statistics
func (m *Monitor) UpdateStats(stats []ContainerStats) {
	m.containers = stats

	// Convert to table rows
	rows := make([]Row, 0, len(stats))
	for _, s := range stats {
		shortID := s.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		cpuStr := fmt.Sprintf("%.1f%%", s.CPU)
		memStr := fmt.Sprintf("%s / %s", FormatBytes(s.Memory), FormatBytes(s.MemoryLimit))
		netStr := fmt.Sprintf("%s / %s", FormatBytes(s.NetRx), FormatBytes(s.NetTx))
		pidsStr := fmt.Sprintf("%d", s.PIDs)

		row := Row{
			shortID,
			s.Name,
			s.Status,
			cpuStr,
			memStr,
			netStr,
			pidsStr,
		}
		rows = append(rows, row)
	}

	m.table.SetRows(rows)
}

// View renders the monitor
func (m Monitor) View() string {
	var b strings.Builder

	// Header
	b.WriteString(TitleStyle.Render("Container Monitor"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Refreshing every %s", m.interval)))
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())

	// Help
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("r: refresh • q: quit • ↑/↓: navigate"))

	return b.String()
}

// SelectedContainer returns the selected container stats
func (m Monitor) SelectedContainer() *ContainerStats {
	row := m.table.SelectedRow()
	if row == nil {
		return nil
	}

	// Find matching container
	for i, c := range m.containers {
		shortID := c.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		if shortID == row[0] {
			return &m.containers[i]
		}
	}

	return nil
}

// Stats represents system-wide stats summary
type Stats struct {
	TotalContainers   int
	RunningContainers int
	StoppedContainers int
	TotalCPU          float64
	TotalMemory       int64
	TotalNetRx        int64
	TotalNetTx        int64
}

// GetSummaryStats returns summary statistics
func (m Monitor) GetSummaryStats() Stats {
	stats := Stats{}

	for _, c := range m.containers {
		stats.TotalContainers++
		switch c.Status {
		case "running":
			stats.RunningContainers++
		case "stopped":
			stats.StoppedContainers++
		}
		stats.TotalCPU += c.CPU
		stats.TotalMemory += c.Memory
		stats.TotalNetRx += c.NetRx
		stats.TotalNetTx += c.NetTx
	}

	return stats
}

// SummaryView returns a summary view
func (m Monitor) SummaryView() string {
	stats := m.GetSummaryStats()

	var b strings.Builder
	b.WriteString(BoxStyle.Render(fmt.Sprintf(
		"Containers: %d (%d running, %d stopped)\n"+
			"Total CPU: %.1f%%\n"+
			"Total Memory: %s\n"+
			"Network: %s rx / %s tx",
		stats.TotalContainers,
		stats.RunningContainers,
		stats.StoppedContainers,
		stats.TotalCPU,
		FormatBytes(stats.TotalMemory),
		FormatBytes(stats.TotalNetRx),
		FormatBytes(stats.TotalNetTx),
	)))

	return b.String()
}
