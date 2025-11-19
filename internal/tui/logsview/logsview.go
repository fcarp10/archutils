package logsview

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/packages"
	"github.com/fcarp10/archutils/internal/scripts"
	"github.com/fcarp10/archutils/internal/tui/helpkeys"
)

var (
	CheckMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	CrossMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("01")).SetString("✗")
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 0)
	spinnerStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

type successInstalledPkg string
type failedInstalledPkg string
type finishedInstallPkgs string
type DisableLogs string
type RunningScript string
type successScript string
type failedScript string

type Model struct {
	progressBar    progress.Model
	spinner        spinner.Model
	numFailedPkgs  int
	numSuccessPkgs int
	packageIndex   int
	packagesLogs   bool
	pkgsSelected   map[int]struct{}
	pkgs           []string
	extensions     bool
	logs           string
}

func (m Model) Init() tea.Cmd {
	return m.progressBar.Init()
}

func NewPackages(pkgsSelected map[int]struct{}, pkgs []string, extensions bool) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner:      s,
		progressBar:  p,
		pkgsSelected: pkgsSelected,
		pkgs:         pkgs,
		extensions:   extensions,
	}
}

func NewScript() Model {
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner: s,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, helpkeys.Keys.Install):
			var pkgNames []string
			for idx := range m.pkgsSelected {
				if idx < len(m.pkgs) {
					pkgNames = append(pkgNames, m.pkgs[idx])
				}
			}
			m.packagesLogs = true
			m.pkgs = pkgNames
			return m, func() tea.Msg {
				return m.installPackage(m.pkgs[m.packageIndex])
			}
		}

	case successInstalledPkg:
		m.logs = fmt.Sprintf("%s %s", CheckMark, strings.Trim(string(msg), "\n"))
		m.numSuccessPkgs++
		return m.selectNextPackages()

	case failedInstalledPkg:
		m.logs = fmt.Sprintf("%s %s", CrossMark, strings.Trim(string(msg), "\n"))
		m.numFailedPkgs++
		return m.selectNextPackages()

	case finishedInstallPkgs:
		m.packageIndex = 0
		m.numFailedPkgs = 0
		m.numSuccessPkgs = 0
		return m, func() tea.Msg { return DisableLogs(msg) }

	case RunningScript:
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg { return runScript(fmt.Sprintf("%v", msg)) })

	case successScript:
		m.logs = fmt.Sprintf("%s %s", CheckMark, msg)
		return m, nil

	case failedScript:
		m.logs = fmt.Sprintf("%s %s", CrossMark, msg)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progressBar.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progressBar = newModel
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) selectNextPackages() (Model, tea.Cmd) {
	prevPkg := m.pkgs[m.packageIndex]
	n := len(m.pkgs)
	if m.packageIndex >= n-1 { // Installation finished
		var doneMsg string
		if m.numFailedPkgs > 0 {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! %d packages installed, %d packages failed.", n-m.numFailedPkgs, m.numFailedPkgs))
		} else {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! All %d packages installed successfully.", n))
		}
		return m, tea.Sequence(
			tea.Printf("%s", m.logs),
			tea.Printf("%s", doneMsg),
			func() tea.Msg { return finishedInstallPkgs(prevPkg) })
	}
	progressCmd := m.progressBar.SetPercent(float64(m.numSuccessPkgs) / float64(n))
	m.packageIndex++ // Move to next pkg
	return m, tea.Batch(
		progressCmd,
		tea.Printf("%s", m.logs), // Print message from previous package
		func() tea.Msg { return m.installPackage(m.pkgs[m.packageIndex]) }, // Install the next package
		m.spinner.Tick,
	)
}

func (m Model) installPackage(pkg string) tea.Msg {
	success, logs := packages.InstallPackage(pkg, m.extensions)
	if success {
		return successInstalledPkg(logs)
	} else {
		return failedInstalledPkg(logs)
	}
}

func runScript(script string) tea.Msg {
	var success bool
	var logs string
	if script == "paru" {
		time.Sleep(time.Second * 1)
		success, logs = scripts.InstallParu()
		if success {
			return successScript(logs)
		} else {
			return failedScript(logs)
		}
	}
	return nil
}

func (m Model) View() string {
	var s string
	if m.packagesLogs {
		n := len(m.pkgs)
		w := lipgloss.Width(fmt.Sprintf("%d", n))
		pkgCount := fmt.Sprintf(" %*d/%*d", w, m.numSuccessPkgs, w, n)

		spin := m.spinner.View() + " "
		progBar := m.progressBar.View()

		pkgName := currentPkgNameStyle.Render(m.pkgs[m.packageIndex])
		info := lipgloss.NewStyle().Render("Installing " + pkgName)

		gap := strings.Repeat(" ", 5)
		s = spin + info + gap + progBar + pkgCount
	} else if m.logs == "" {
		spin := m.spinner.View() + " "
		s = spin + "Running script, please wait..."
	} else {
		s = m.logs
	}
	return s
}
