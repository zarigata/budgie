package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
)

// ProgressBar represents a progress bar component
type ProgressBar struct {
	progress progress.Model
	title    string
	percent  float64
	width    int
}

// NewProgressBar creates a new progress bar
func NewProgressBar(title string) ProgressBar {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return ProgressBar{
		progress: p,
		title:    title,
		percent:  0,
		width:    40,
	}
}

// SetPercent sets the progress percentage (0.0 to 1.0)
func (p *ProgressBar) SetPercent(percent float64) {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	p.percent = percent
}

// SetWidth sets the progress bar width
func (p *ProgressBar) SetWidth(width int) {
	p.width = width
	p.progress.Width = width
}

// Init initializes the progress bar
func (p ProgressBar) Init() tea.Cmd {
	return nil
}

// Update handles progress bar updates
func (p ProgressBar) Update(msg tea.Msg) (ProgressBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width - 10
		if p.width > 80 {
			p.width = 80
		}
		if p.width < 20 {
			p.width = 20
		}
		p.progress.Width = p.width
	case progress.FrameMsg:
		progressModel, cmd := p.progress.Update(msg)
		p.progress = progressModel.(progress.Model)
		return p, cmd
	}
	return p, nil
}

// View renders the progress bar
func (p ProgressBar) View() string {
	var b strings.Builder

	if p.title != "" {
		b.WriteString(p.title)
		b.WriteString("\n")
	}

	b.WriteString(p.progress.ViewAs(p.percent))
	b.WriteString(fmt.Sprintf(" %.0f%%", p.percent*100))

	return b.String()
}

// Percent returns the current percentage
func (p ProgressBar) Percent() float64 {
	return p.percent
}

// TransferProgress shows transfer progress with speed and ETA
type TransferProgress struct {
	bar         ProgressBar
	totalBytes  int64
	sentBytes   int64
	startTime   int64
	speedBps    int64
	description string
}

// NewTransferProgress creates a new transfer progress tracker
func NewTransferProgress(description string, totalBytes int64) TransferProgress {
	return TransferProgress{
		bar:         NewProgressBar(description),
		totalBytes:  totalBytes,
		description: description,
	}
}

// Update updates the transfer progress
func (t *TransferProgress) Update(sentBytes int64, elapsedMs int64) {
	t.sentBytes = sentBytes

	if t.totalBytes > 0 {
		t.bar.SetPercent(float64(sentBytes) / float64(t.totalBytes))
	}

	if elapsedMs > 0 {
		t.speedBps = (sentBytes * 1000) / elapsedMs
	}
}

// View renders the transfer progress
func (t TransferProgress) View() string {
	var b strings.Builder

	b.WriteString(t.bar.View())
	b.WriteString("\n")

	// Speed and transferred
	speedStr := FormatBytes(t.speedBps) + "/s"
	transferredStr := fmt.Sprintf("%s / %s", FormatBytes(t.sentBytes), FormatBytes(t.totalBytes))

	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("%s  |  %s", transferredStr, speedStr)))

	// ETA
	if t.speedBps > 0 && t.sentBytes < t.totalBytes {
		remaining := t.totalBytes - t.sentBytes
		etaSeconds := remaining / t.speedBps
		if etaSeconds < 60 {
			b.WriteString(fmt.Sprintf("  |  ETA: %ds", etaSeconds))
		} else if etaSeconds < 3600 {
			b.WriteString(fmt.Sprintf("  |  ETA: %dm %ds", etaSeconds/60, etaSeconds%60))
		} else {
			b.WriteString(fmt.Sprintf("  |  ETA: %dh %dm", etaSeconds/3600, (etaSeconds%3600)/60))
		}
	}

	return b.String()
}

// MultiProgress shows multiple progress bars
type MultiProgress struct {
	bars   []ProgressBar
	width  int
	height int
}

// NewMultiProgress creates a multi-progress tracker
func NewMultiProgress() MultiProgress {
	return MultiProgress{
		bars: make([]ProgressBar, 0),
	}
}

// AddBar adds a progress bar
func (m *MultiProgress) AddBar(title string) int {
	bar := NewProgressBar(title)
	m.bars = append(m.bars, bar)
	return len(m.bars) - 1
}

// SetPercent sets a specific bar's percentage
func (m *MultiProgress) SetPercent(index int, percent float64) {
	if index >= 0 && index < len(m.bars) {
		m.bars[index].SetPercent(percent)
	}
}

// Init initializes multi-progress
func (m MultiProgress) Init() tea.Cmd {
	return nil
}

// Update handles updates
func (m MultiProgress) Update(msg tea.Msg) (MultiProgress, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		for i := range m.bars {
			m.bars[i].SetWidth(msg.Width - 10)
		}
	}

	var cmds []tea.Cmd
	for i := range m.bars {
		var cmd tea.Cmd
		m.bars[i], cmd = m.bars[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders all progress bars
func (m MultiProgress) View() string {
	var b strings.Builder

	for i, bar := range m.bars {
		b.WriteString(bar.View())
		if i < len(m.bars)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

// OverallPercent returns the overall progress
func (m MultiProgress) OverallPercent() float64 {
	if len(m.bars) == 0 {
		return 0
	}

	total := 0.0
	for _, bar := range m.bars {
		total += bar.Percent()
	}

	return total / float64(len(m.bars))
}
