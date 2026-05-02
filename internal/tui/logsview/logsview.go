package logsview

import (
	"fmt"
	"log"
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
	errorCountStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("01"))
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
type CancelInstall struct{}
type SudoValidated struct{ err error }

const (
	ScriptParu ScriptType = iota
	ScriptAutologin
	ScriptPasswordlessSSH
	ScriptPasswordlessSudo
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
	itemLogs        bool
	itemNames       []string
	itemType        ItemsInstallType
	logs            string
	installer       scripts.Installer
	failedItemLogs  []string
	cancelRequested bool
	pendingScript   ScriptType
	validatingSudo  bool
	scriptRunning   bool
}

func (m Model) Init() tea.Cmd {
	return m.progressBar.Init()
}

func (m Model) IsActive() bool {
	return m.validatingSudo || m.scriptRunning || m.itemLogs
}

func NewItems(itemNames []string, installer scripts.Installer) Model {
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
		itemNames:   itemNames,
		installer:   installer,
	}
}

func NewScript(installer scripts.Installer) Model {
	s := spinner.New()
	s.Style = spinnerStyle
	return Model{
		spinner:   s,
		installer: installer,
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
		m.itemLogs = true
		m.itemType = ItemsInstallType(msg)
		if ItemsInstallType(msg) == InstallPackages {
			m.validatingSudo = true
			return m, tea.ExecProcess(m.installer.SudoValidateCmd(), func(err error) tea.Msg {
				return SudoValidated{err: err}
			})
		}
		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg { return m.installItem(m.itemType) },
		)

	case SudoValidated:
		m.validatingSudo = false
		if msg.err != nil {
			m.itemLogs = false
			return m, func() tea.Msg { return DisableLogs("Sudo authentication failed: password is required") }
		}
		if m.itemLogs {
			return m, tea.Batch(
				m.spinner.Tick,
				func() tea.Msg { return m.installItem(m.itemType) },
			)
		}
		m.scriptRunning = true
		installer := m.installer
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg { return runScript(installer, m.pendingScript) })

	case successInstalledItem:
		m.logs = fmt.Sprintf("%s %s", CheckMark, strings.Trim(string(msg), "\n"))
		m.successItemsNum++
		return m.selectNextItem(m.itemType)

	case failedInstalledItem:
		m.logs = fmt.Sprintf("%s %s", CrossMark, strings.Trim(string(msg), "\n"))
		m.failedItemsNum++
		m.failedItemLogs = append(m.failedItemLogs, m.logs)
		return m.selectNextItem(m.itemType)

	case finishedInstallItems:
		var summary string
		if len(m.failedItemLogs) > 0 {
			summary = "\nFailed items:\n"
			for _, errLog := range m.failedItemLogs {
				summary += "  " + errLog + "\n"
			}
		}
		m.itemIndex = 0
		m.failedItemsNum = 0
		m.successItemsNum = 0
		m.cancelRequested = false
		m.failedItemLogs = nil
		return m, func() tea.Msg { return DisableLogs(summary) }

	case CancelInstall:
		m.cancelRequested = true
		return m, nil

	case RunningScript:
		m.pendingScript = ScriptType(msg)
		m.validatingSudo = true
		return m, tea.ExecProcess(m.installer.SudoValidateCmd(), func(err error) tea.Msg {
			return SudoValidated{err: err}
		})

	case successScript:
		m.scriptRunning = false
		m.logs = fmt.Sprintf("%s %s", CheckMark, msg)
		return m, nil

	case failedScript:
		m.scriptRunning = false
		m.logs = fmt.Sprintf("%s %s", CrossMark, msg)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progressBar.Update(msg)
		if pm, ok := newModel.(progress.Model); ok {
			m.progressBar = pm
		} else {
			log.Printf("warning: progress bar update returned unexpected type %T", newModel)
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) selectNextItem(itemsType ItemsInstallType) (Model, tea.Cmd) {
	prevPkg := m.itemNames[m.itemIndex]
	n := len(m.itemNames)

	isFinished := m.itemIndex >= n-1
	isCancelled := m.cancelRequested

	if isCancelled || isFinished {
		var doneMsg string
		if isCancelled {
			doneMsg = doneStyle.Render(fmt.Sprintf("Cancelled! %d/%d items installed.", m.successItemsNum, n))
		} else if m.failedItemsNum > 0 {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! %d items installed, %d items failed.", m.successItemsNum, m.failedItemsNum))
		} else {
			doneMsg = doneStyle.Render(fmt.Sprintf("Done! All %d items installed successfully.", n))
		}
		return m, tea.Sequence(
			tea.Printf("%s", m.logs),
			tea.Printf("%s", doneMsg),
			func() tea.Msg { return finishedInstallItems(prevPkg) })
	}
	progressCmd := m.progressBar.SetPercent(float64(m.successItemsNum) / float64(n))
	m.itemIndex++
	return m, tea.Batch(
		progressCmd,
		tea.Printf("%s", m.logs),
		func() tea.Msg { return m.installItem(ItemsInstallType(itemsType)) },
		m.spinner.Tick,
	)
}

func (m Model) installItem(itemsType ItemsInstallType) tea.Msg {
	var success bool
	var logs string
	switch itemsType {
	case InstallPackages:
		success, logs = m.installer.InstallPackage(m.itemNames[m.itemIndex])
	case InstallExtensions:
		success, logs = m.installer.InstallVSCodeExtension(m.itemNames[m.itemIndex])
	}
	if success {
		return successInstalledItem(logs)
	} else {
		return failedInstalledItem(logs)
	}
}

func runScript(installer scripts.Installer, script ScriptType) tea.Msg {
	if installer == nil {
		return failedScript("Installer not available")
	}
	var success bool
	var logs string
	switch script {
	case ScriptParu:
		success, logs = installer.InstallParu()
	case ScriptAutologin:
		success, logs = installer.EnableAutologin()
	case ScriptPasswordlessSSH:
		success, logs = installer.EnablePasswordlessSSH()
	case ScriptPasswordlessSudo:
		success, logs = installer.EnablePasswordlessSudo()
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
	if m.validatingSudo {
		spin := m.spinner.View() + " "
		s = spin + "Authenticating with sudo, please enter your password..."
	} else if m.itemLogs {
		n := len(m.itemNames)
		w := lipgloss.Width(fmt.Sprintf("%d", n))
		itemCount := fmt.Sprintf(" %*d/%*d", w, m.successItemsNum, w, n)

		spin := m.spinner.View() + " "
		progBar := m.progressBar.View()

		itemName := currentPkgNameStyle.Render(m.itemNames[m.itemIndex])
		info := lipgloss.NewStyle().Render("Installing " + itemName)

		gap := strings.Repeat(" ", 5)
		s = spin + info + gap + progBar + itemCount

		if m.failedItemsNum > 0 {
			failInfo := errorCountStyle.Render(fmt.Sprintf(" (%d failed)", m.failedItemsNum))
			s += failInfo
		}
	} else if m.logs == "" {
		spin := m.spinner.View() + " "
		s = spin + "Running script, please wait..."
	} else {
		s = m.logs
	}
	return s
}
