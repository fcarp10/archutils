package logsview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/scripts"
)

var (
	CheckMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	CrossMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("01")).SetString("✗")
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 0)
	spinnerStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

type DisableLogs string
type ScriptType int
type RunningScript ScriptType
type successScript string
type failedScript string
type ItemsInstallType int
type InstallItems ItemsInstallType
type successInstalledItem string
type failedInstalledItem string
type finishedInstallItems string

const (
	ScriptParu ScriptType = iota
	ScriptAutologin
	ScriptPasswordlessSSH
)

const (
	InstallPackages ItemsInstallType = iota
	InstallExtensions
)

type Model struct {
	progressBar     progress.Model
	spinner         spinner.Model
	failedItemsNum  int
	successItemsNum int
	itemIndex       int
	itemsLogs       bool
	itemsNames      []string
	itemsType       ItemsInstallType
	logs            string
}

func (m Model) Init() tea.Cmd {
	return m.progressBar.Init()
}

func NewItems(itemsNames []string) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner:     s,
		progressBar: p,
		itemsNames:  itemsNames,
	}
}

func NewScript() Model {
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner: s,
	}
}

func NewInfo(info string) Model {
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		logs: info,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {

	case InstallItems:
		m.itemsLogs = true
		m.itemsType = ItemsInstallType(msg)
		return m, func() tea.Msg { return m.installItem(m.itemsType) }

	case successInstalledItem:
		m.logs = fmt.Sprintf("%s %s", CheckMark, strings.Trim(string(msg), "\n"))
		m.successItemsNum++
		return m.selectNextItem(m.itemsType)

	case failedInstalledItem:
		m.logs = fmt.Sprintf("%s %s", CrossMark, strings.Trim(string(msg), "\n"))
		m.failedItemsNum++
		return m.selectNextItem(m.itemsType)

	case finishedInstallItems:
		m.itemIndex = 0
		m.failedItemsNum = 0
		m.successItemsNum = 0
		return m, func() tea.Msg { return DisableLogs(msg) }

	case RunningScript:
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg { return runScript(ScriptType(msg)) })

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

func (m Model) selectNextItem(itemsType ItemsInstallType) (Model, tea.Cmd) {
	prevPkg := m.itemsNames[m.itemIndex]
	n := len(m.itemsNames)
	if m.itemIndex >= n-1 { // Installation finished
		var doneMsg string
		if m.failedItemsNum > 0 {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! %d items installed, %d items failed.", n-m.failedItemsNum, m.failedItemsNum))
		} else {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! All %d items installed successfully.", n))
		}
		return m, tea.Sequence(
			tea.Printf("%s", m.logs),
			tea.Printf("%s", doneMsg),
			func() tea.Msg { return finishedInstallItems(prevPkg) })
	}
	progressCmd := m.progressBar.SetPercent(float64(m.successItemsNum) / float64(n))
	m.itemIndex++ // Move to next item
	return m, tea.Batch(
		progressCmd,
		tea.Printf("%s", m.logs), // Print message from previous item
		func() tea.Msg { return m.installItem(ItemsInstallType(itemsType)) }, // Install the next item
		m.spinner.Tick,
	)
}

func (m Model) installItem(itemsType ItemsInstallType) tea.Msg {
	var success bool
	var logs string
	switch itemsType {
	case InstallPackages:
		success, logs = scripts.InstallPackage(m.itemsNames[m.itemIndex])
	case InstallExtensions:
		success, logs = scripts.InstallVSCodeExtension(m.itemsNames[m.itemIndex])
	}
	if success {
		return successInstalledItem(logs)
	} else {
		return failedInstalledItem(logs)
	}
}

func runScript(script ScriptType) tea.Msg {
	var success bool
	var logs string
	switch script {
	case ScriptParu:
		success, logs = scripts.InstallParu()
	case ScriptAutologin:
		success, logs = scripts.EnableAutologin()
	case ScriptPasswordlessSSH:
		success, logs = scripts.EnablePasswordlessSSH()
	default:
	}
	if success {
		return successScript(logs)
	} else {
		return failedScript(logs)
	}
}

func (m Model) View() string {
	var s string
	if m.itemsLogs {
		n := len(m.itemsNames)
		w := lipgloss.Width(fmt.Sprintf("%d", n))
		itemCount := fmt.Sprintf(" %*d/%*d", w, m.successItemsNum, w, n)

		spin := m.spinner.View() + " "
		progBar := m.progressBar.View()

		itemName := currentPkgNameStyle.Render(m.itemsNames[m.itemIndex])
		info := lipgloss.NewStyle().Render("Installing " + itemName)

		gap := strings.Repeat(" ", 5)
		s = spin + info + gap + progBar + itemCount
	} else if m.logs == "" {
		spin := m.spinner.View() + " "
		s = spin + "Running script, please wait..."
	} else {
		s = m.logs
	}
	return s
}
