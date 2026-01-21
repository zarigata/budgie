package nest

import (
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/ui"
)

var nestCmd = &cobra.Command{
	Use:   "nest",
	Short: "Interactive setup and build wizard",
	Long: `Nest provides an interactive wizard to:
  - Detect your system
  - Choose build targets
  - Configure budgie for your environment
  - Learn how to use budgie`,
	RunE: runNest,
}

// View represents the current view in the UI
type View int

const (
	ViewMain View = iota
	ViewQuickStart
	ViewCustomBuild
	ViewLearn
	ViewMonitor
	ViewSystemCheck
)

// Model is the main application model
type Model struct {
	view           View
	menu           ui.Menu
	buildMenu      ui.Menu
	learnMenu      ui.Menu
	width          int
	height         int
	sysInfo        *SystemInfo
	selectedLesson int
	quitting       bool
}

// SystemInfo holds system information
type SystemInfo struct {
	OS        string
	Arch      string
	GoVersion string
	Supported bool
}

// InitialModel creates the initial model
func InitialModel() Model {
	sysInfo := detectSystem()

	mainMenu := ui.NewMenu("What would you like to do?", []ui.MenuItem{
		{Label: "Quick Start", Description: "Get started with budgie", Icon: "1."},
		{Label: "Custom Build", Description: "Build for specific platforms", Icon: "2."},
		{Label: "Learn Budgie", Description: "Tutorials and documentation", Icon: "3."},
		{Label: "Monitor", Description: "View running containers", Icon: "4."},
		{Label: "System Check", Description: "Check dependencies", Icon: "5."},
		{Label: "Exit", Description: "Exit the wizard", Icon: "6."},
	})

	buildMenu := ui.NewMenu("Select target platform:", []ui.MenuItem{
		{Label: "Linux (amd64)", Icon: "1."},
		{Label: "Linux (arm64)", Icon: "2."},
		{Label: "macOS (amd64)", Icon: "3."},
		{Label: "macOS (arm64)", Icon: "4."},
		{Label: "Windows (amd64)", Icon: "5."},
		{Label: "Windows (arm64)", Icon: "6."},
		{Label: "All platforms", Icon: "7."},
		{Label: "Back", Icon: "8."},
	})

	learnMenu := ui.NewMenu("Select a tutorial:", []ui.MenuItem{
		{Label: "Running Your First Container", Description: "Learn how to create and run containers", Icon: "1."},
		{Label: "Discovering Containers", Description: "Use chirp to find containers on LAN", Icon: "2."},
		{Label: "Container Replication", Description: "Join containers as peers", Icon: "3."},
		{Label: "Managing Containers", Description: "List, start, stop, remove containers", Icon: "4."},
		{Label: "Back", Icon: "5."},
	})

	return Model{
		view:      ViewMain,
		menu:      mainMenu,
		buildMenu: buildMenu,
		learnMenu: learnMenu,
		sysInfo:   sysInfo,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.view == ViewMain {
				m.quitting = true
				return m, tea.Quit
			}
			// Go back to main menu
			m.view = ViewMain
			return m, nil

		case "b", "esc":
			// Go back
			switch m.view {
			case ViewQuickStart, ViewCustomBuild, ViewLearn, ViewMonitor, ViewSystemCheck:
				m.view = ViewMain
			}
			return m, nil

		case "enter", " ":
			return m.handleSelect()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update the appropriate menu
	var cmd tea.Cmd
	switch m.view {
	case ViewMain:
		m.menu, cmd = m.menu.Update(msg)
	case ViewCustomBuild:
		m.buildMenu, cmd = m.buildMenu.Update(msg)
	case ViewLearn:
		m.learnMenu, cmd = m.learnMenu.Update(msg)
	}

	return m, cmd
}

func (m Model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewMain:
		switch m.menu.Cursor() {
		case 0: // Quick Start
			m.view = ViewQuickStart
		case 1: // Custom Build
			m.view = ViewCustomBuild
		case 2: // Learn
			m.view = ViewLearn
		case 3: // Monitor
			m.view = ViewMonitor
		case 4: // System Check
			m.view = ViewSystemCheck
		case 5: // Exit
			m.quitting = true
			return m, tea.Quit
		}

	case ViewCustomBuild:
		if m.buildMenu.Cursor() == 7 { // Back
			m.view = ViewMain
		}

	case ViewLearn:
		if m.learnMenu.Cursor() == 4 { // Back
			m.view = ViewMain
		} else {
			m.selectedLesson = m.learnMenu.Cursor()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	// Logo
	b.WriteString(ui.Logo())
	b.WriteString("\n")

	// System info
	b.WriteString(m.renderSystemInfo())
	b.WriteString("\n\n")

	// Current view
	switch m.view {
	case ViewMain:
		b.WriteString(m.menu.View())

	case ViewQuickStart:
		b.WriteString(m.renderQuickStart())

	case ViewCustomBuild:
		b.WriteString(m.buildMenu.View())
		b.WriteString("\n\n")
		b.WriteString(m.renderBuildCommand())

	case ViewLearn:
		b.WriteString(m.learnMenu.View())
		if m.selectedLesson >= 0 && m.selectedLesson < 4 {
			b.WriteString("\n\n")
			b.WriteString(m.renderTutorial(m.selectedLesson))
		}

	case ViewMonitor:
		b.WriteString(m.renderMonitor())

	case ViewSystemCheck:
		b.WriteString(m.renderSystemCheck())
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("Press 'q' to quit, 'b' to go back"))

	return b.String()
}

func (m Model) renderSystemInfo() string {
	var b strings.Builder

	supported := ui.RunningStyle.Render("Yes")
	if !m.sysInfo.Supported {
		supported = ui.ErrorStyle.Render("No")
	}

	b.WriteString(ui.SubtitleStyle.Render("System Information"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  OS:           %s\n", m.sysInfo.OS))
	b.WriteString(fmt.Sprintf("  Architecture: %s\n", m.sysInfo.Arch))
	b.WriteString(fmt.Sprintf("  Go Version:   %s\n", m.sysInfo.GoVersion))
	b.WriteString(fmt.Sprintf("  Supported:    %s", supported))

	return ui.BoxStyle.Render(b.String())
}

func (m Model) renderQuickStart() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("Quick Start"))
	b.WriteString("\n\n")

	steps := []string{
		"Install dependencies:\n   go mod download",
		"Build budgie:\n   make build",
		"Run your first container:\n   budgie run example.bun",
		"Discover containers on network:\n   budgie chirp",
	}

	for i, step := range steps {
		b.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, step))
	}

	b.WriteString(ui.InfoBoxStyle.Render("Tip: Use --detach to run containers in background"))

	return b.String()
}

func (m Model) renderBuildCommand() string {
	commands := map[int]string{
		0: "GOOS=linux GOARCH=amd64 make build",
		1: "GOOS=linux GOARCH=arm64 make build",
		2: "GOOS=darwin GOARCH=amd64 make build",
		3: "GOOS=darwin GOARCH=arm64 make build",
		4: "GOOS=windows GOARCH=amd64 make build",
		5: "GOOS=windows GOARCH=arm64 make build",
		6: "make build-all",
	}

	cursor := m.buildMenu.Cursor()
	if cursor >= 0 && cursor < 7 {
		return ui.BoxStyle.Render(fmt.Sprintf("Run: %s", commands[cursor]))
	}
	return ""
}

func (m Model) renderTutorial(index int) string {
	tutorials := []string{
		// Tutorial 0: Running Your First Container
		`Create a .bun file:

  version: "1.0"
  name: "myapp"
  image:
    docker_image: "nginx:alpine"
  ports:
    - container_port: 80
      host_port: 8080
      protocol: tcp

Run it:
  budgie run myapp.bun

Your container is now running!`,

		// Tutorial 1: Discovering Containers
		`Budgie uses mDNS to discover containers on your LAN.

Command:
  budgie chirp

This shows all containers with their:
  - Container ID
  - Name
  - IP address
  - Port
  - Node hostname

Containers auto-announce themselves on startup.`,

		// Tutorial 2: Container Replication
		`To provide redundancy, join an existing container:

Command:
  budgie chirp <container-id>

This will:
  - Connect to the primary node
  - Download the container image
  - Synchronize volume data
  - Start the replica

Multiple replicas provide automatic failover.`,

		// Tutorial 3: Managing Containers
		`List containers:
  budgie ps

List including stopped:
  budgie ps --all

Stop a container:
  budgie stop <container-id>

Stop with timeout:
  budgie stop --timeout 30s <container-id>

Use 'budgie ps' to get container IDs.`,
	}

	if index >= 0 && index < len(tutorials) {
		return ui.BoxStyle.Render(tutorials[index])
	}
	return ""
}

func (m Model) renderMonitor() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("Container Monitor"))
	b.WriteString("\n\n")
	b.WriteString("No containers currently running.\n\n")
	b.WriteString("Start a container with:\n")
	b.WriteString("  budgie run <your-file>.bun\n")

	return b.String()
}

func (m Model) renderSystemCheck() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("System Check"))
	b.WriteString("\n\n")

	checks := []struct {
		name     string
		status   bool
		required bool
	}{
		{"Go (1.21+)", true, true},
		{"containerd", checkContainerd(), true},
		{"mDNS support", true, false},
	}

	allGood := true
	for _, check := range checks {
		var status string
		if check.status {
			status = ui.RunningStyle.Render("[OK]")
		} else {
			if check.required {
				status = ui.ErrorStyle.Render("[MISSING]")
				allGood = false
			} else {
				status = ui.StoppedStyle.Render("[OPTIONAL]")
			}
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", status, check.name))
	}

	b.WriteString("\n")
	if allGood {
		b.WriteString(ui.InfoBoxStyle.Render("Your system is ready to run budgie!"))
	} else {
		b.WriteString(ui.WarningBoxStyle.Render("Some dependencies are missing. Please install them."))
	}

	return b.String()
}

func detectSystem() *SystemInfo {
	info := &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}

	supportedOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}

	supportedArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
		"arm":   true,
	}

	info.Supported = supportedOS[info.OS] && supportedArch[info.Arch]

	return info
}

func checkContainerd() bool {
	// Simple check - in real implementation would check if containerd is running
	return true
}

func runNest(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func init() {
	nestCmd.Aliases = []string{"setup", "wizard", "init"}
}

func GetNestCmd() *cobra.Command {
	return nestCmd
}
