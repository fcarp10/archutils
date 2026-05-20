package logsview

import (
	"os/exec"
	"testing"
)

// mockScriptInstaller implements scripts.Installer with minimal stubs for logsview testing.
type mockScriptInstaller struct {
	installPkg     func(string) (bool, string)
	installExt     func(string) (bool, string)
	installParu    func() (bool, string)
	autologin      func() (bool, string)
	passwordless   func() (bool, string)
	sudo           func() (bool, string)
	addUserToWheel func() (bool, string)
}

func (m mockScriptInstaller) InstallPackage(pkg string) (bool, string) {
	if m.installPkg != nil {
		return m.installPkg(pkg)
	}
	return true, pkg + ": installed"
}

func (m mockScriptInstaller) InstallParu() (bool, string) {
	if m.installParu != nil {
		return m.installParu()
	}
	return true, "paru installed"
}

func (m mockScriptInstaller) InstallVSCodeExtension(ext string) (bool, string) {
	if m.installExt != nil {
		return m.installExt(ext)
	}
	return true, ext + ": installed"
}

func (m mockScriptInstaller) EnableAutologin() (bool, string) {
	if m.autologin != nil {
		return m.autologin()
	}
	return true, "autologin enabled"
}

func (m mockScriptInstaller) EnablePasswordlessSSH() (bool, string) {
	if m.passwordless != nil {
		return m.passwordless()
	}
	return true, "ssh configured"
}

func (m mockScriptInstaller) EnablePasswordlessSudo() (bool, string) {
	if m.sudo != nil {
		return m.sudo()
	}
	return true, "sudo configured"
}

func (m mockScriptInstaller) AddUserToWheel() (bool, string) {
	if m.addUserToWheel != nil {
		return m.addUserToWheel()
	}
	return true, "user added to wheel"
}

func (m mockScriptInstaller) GetPackageDescription(item string) string {
	return "desc of " + item
}

func (m mockScriptInstaller) GetExtensionDescription(ext string) string {
	return "desc of " + ext
}

func (m mockScriptInstaller) CheckParuInstalled() (bool, string)   { return true, "" }
func (m mockScriptInstaller) IsPackageInstalled(pkg string) bool   { return false }
func (m mockScriptInstaller) IsExtensionInstalled(ext string) bool { return false }
func (m mockScriptInstaller) SudoValidateCmd() *exec.Cmd {
	return exec.Command("true")
}
func (m mockScriptInstaller) GetInstalledPackages() map[string]string { return nil }

func TestNewInfo(t *testing.T) {
	m := NewInfo("test message")
	if m.logs != "test message" {
		t.Errorf("expected logs 'test message', got %q", m.logs)
	}
	if m.itemLogs {
		t.Error("expected itemLogs false for info model")
	}
}

func TestNewItems(t *testing.T) {
	m := NewItems([]string{"pkg1", "pkg2"}, mockScriptInstaller{})
	if len(m.itemNames) != 2 {
		t.Errorf("expected 2 item names, got %d", len(m.itemNames))
	}
}

func TestInstallItems_Packages(t *testing.T) {
	m := NewItems([]string{"pkg1"}, mockScriptInstaller{})
	m, cmd := m.Update(InstallItems(InstallPackages))
	if !m.validatingSudo {
		t.Error("expected validatingSudo true for packages")
	}
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestInstallItems_Extensions(t *testing.T) {
	m := NewItems([]string{"ext1"}, mockScriptInstaller{})
	m, cmd := m.Update(InstallItems(InstallExtensions))
	if m.validatingSudo {
		t.Error("expected validatingSudo false for extensions")
	}
	if !m.itemLogs {
		t.Error("expected itemLogs true for extensions")
	}
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestSudoValidated_Success(t *testing.T) {
	m := NewItems([]string{"pkg1"}, mockScriptInstaller{})
	m.validatingSudo = true

	m, cmd := m.Update(SudoValidated{err: nil})
	if m.validatingSudo {
		t.Error("expected validatingSudo to be cleared")
	}
	if cmd == nil {
		t.Error("expected non-nil command after sudo validated")
	}
}

func TestSudoValidated_Failure(t *testing.T) {
	m := NewItems([]string{"pkg1"}, mockScriptInstaller{})
	m.validatingSudo = true
	m.itemLogs = true

	m, cmd := m.Update(SudoValidated{err: exec.ErrNotFound})
	if m.validatingSudo {
		t.Error("expected validatingSudo to be cleared")
	}
	if cmd == nil {
		t.Error("expected non-nil command after sudo failure")
	}
}

func TestCancelInstall(t *testing.T) {
	m := NewItems([]string{"pkg1", "pkg2"}, mockScriptInstaller{})
	m.itemLogs = true
	m.itemIndex = 0

	m, _ = m.Update(CancelInstall{})
	if !m.cancelRequested {
		t.Error("expected cancelRequested to be true")
	}
}

func TestSuccessInstalledItem(t *testing.T) {
	m := NewItems([]string{"pkg1", "pkg2"}, mockScriptInstaller{})
	m.itemIndex = 0
	m.itemLogs = true

	m, cmd := m.Update(successInstalledItem("pkg1 installed"))
	if m.successItemsNum != 1 {
		t.Errorf("expected 1 success, got %d", m.successItemsNum)
	}
	if cmd == nil {
		t.Error("expected non-nil command (selectNextItem)")
	}
}

func TestFailedInstalledItem(t *testing.T) {
	m := NewItems([]string{"pkg1", "pkg2"}, mockScriptInstaller{})
	m.itemIndex = 0
	m.itemLogs = true

	m, cmd := m.Update(failedInstalledItem("pkg1 failed"))
	if m.failedItemsNum != 1 {
		t.Errorf("expected 1 failure, got %d", m.failedItemsNum)
	}
	if len(m.failedItemLogs) != 1 {
		t.Errorf("expected 1 failed item log, got %d", len(m.failedItemLogs))
	}
	if cmd == nil {
		t.Error("expected non-nil command (selectNextItem)")
	}
}

func TestFinishedInstallItems(t *testing.T) {
	m := NewItems([]string{"pkg1"}, mockScriptInstaller{})
	m.successItemsNum = 1
	m.itemIndex = 0

	// Simulate finishing through selectNextItem by advancing past last item
	msg := finishedInstallItems("pkg1")
	m, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected non-nil command")
	}
	if m.successItemsNum != 0 {
		t.Errorf("expected success count reset to 0, got %d", m.successItemsNum)
	}
	if m.failedItemsNum != 0 {
		t.Errorf("expected failure count reset to 0, got %d", m.failedItemsNum)
	}
}

func TestRunningScript(t *testing.T) {
	m := NewScript(mockScriptInstaller{})
	m, cmd := m.Update(RunningScript(ScriptParu))
	if !m.validatingSudo {
		t.Error("expected validatingSudo true for script")
	}
	if m.pendingScript != ScriptParu {
		t.Errorf("expected pendingScript ScriptParu, got %d", m.pendingScript)
	}
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestSuccessScript(t *testing.T) {
	m := NewScript(mockScriptInstaller{})
	m, _ = m.Update(successScript("script done"))
	if m.scriptRunning {
		t.Error("expected scriptRunning to be false")
	}
}

func TestFailedScript(t *testing.T) {
	m := NewScript(mockScriptInstaller{})
	m, _ = m.Update(failedScript("script failed"))
	if m.scriptRunning {
		t.Error("expected scriptRunning to be false")
	}
}

func TestViewModes(t *testing.T) {
	// Info view
	m := NewInfo("test message")
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view for info")
	}

	// Validating sudo view
	m = NewScript(mockScriptInstaller{})
	m.validatingSudo = true
	v = m.View()
	if v == "" {
		t.Error("expected non-empty view for sudo validation")
	}

	// Installing view
	m = NewItems([]string{"pkg1"}, mockScriptInstaller{})
	m.itemLogs = true
	m.itemIndex = 0
	v = m.View()
	if v == "" {
		t.Error("expected non-empty view for installing")
	}
}
