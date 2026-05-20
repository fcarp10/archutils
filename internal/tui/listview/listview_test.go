package listview

import (
	"os/exec"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fcarp10/archutils/internal/config"
)

// mockInstaller implements scripts.Installer for use in tests.
type mockInstaller struct {
	installedPkgs    map[string]string
	packageInstalled map[string]bool
}

func (m mockInstaller) InstallPackage(pkg string) (bool, string) {
	return true, pkg + ": installed"
}
func (m mockInstaller) InstallParu() (bool, string) { return true, "paru installed" }
func (m mockInstaller) InstallVSCodeExtension(ext string) (bool, string) {
	return true, ext + ": installed"
}
func (m mockInstaller) EnableAutologin() (bool, string)           { return true, "autologin enabled" }
func (m mockInstaller) EnablePasswordlessSSH() (bool, string)     { return true, "ssh configured" }
func (m mockInstaller) EnablePasswordlessSudo() (bool, string)    { return true, "sudo configured" }
func (m mockInstaller) AddUserToWheel() (bool, string)            { return true, "user added to wheel" }
func (m mockInstaller) WheelGroupCmd() *exec.Cmd                  { return exec.Command("true") }
func (m mockInstaller) GetPackageDescription(item string) string  { return "description of " + item }
func (m mockInstaller) GetExtensionDescription(ext string) string { return "description of " + ext }
func (m mockInstaller) CheckParuInstalled() (bool, string)        { return true, "" }
func (m mockInstaller) IsPackageInstalled(pkg string) bool {
	if m.packageInstalled == nil {
		return false
	}
	return m.packageInstalled[pkg]
}
func (m mockInstaller) IsExtensionInstalled(ext string) bool { return false }
func (m mockInstaller) SudoValidateCmd() *exec.Cmd {
	return exec.Command("true")
}
func (m mockInstaller) GetInstalledPackages() map[string]string {
	if m.installedPkgs == nil {
		return map[string]string{}
	}
	return m.installedPkgs
}

func TestNew(t *testing.T) {
	m := New(mockInstaller{})

	if m.currentStage != stageMenu {
		t.Errorf("expected stageMenu (%d), got %d", stageMenu, m.currentStage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
	if m.installer == nil {
		t.Error("installer should not be nil")
	}
}

func TestSelectionCount_MenuStage(t *testing.T) {
	m := New(mockInstaller{})
	selected, total := m.SelectionCount()
	if selected != -1 || total != -1 {
		t.Errorf("expected (-1, -1) for menu stage, got (%d, %d)", selected, total)
	}
}

func TestSelectionCount_ItemsStage(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems
	m.itemNames = []string{"pkg1", "pkg2", "pkg3"}
	m.selectedItems = map[int]struct{}{0: {}, 2: {}}

	selected, total := m.SelectionCount()
	if selected != 2 {
		t.Errorf("expected 2 selected, got %d", selected)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
}

func TestInitializeSelection(t *testing.T) {
	items := []string{"# disabled-pkg", "enabled-pkg", "# another-disabled", "another-enabled"}
	processed, selected := initializeSelection(items)

	expectedProcessed := []string{"disabled-pkg", "enabled-pkg", "another-disabled", "another-enabled"}
	for i, want := range expectedProcessed {
		if processed[i] != want {
			t.Errorf("item %d: expected %q, got %q", i, want, processed[i])
		}
	}

	// #-prefixed items should NOT be selected
	if _, ok := selected[0]; ok {
		t.Error("expected item 0 (# disabled-pkg) to NOT be selected")
	}
	if _, ok := selected[1]; !ok {
		t.Error("expected item 1 (enabled-pkg) to be selected")
	}
	if _, ok := selected[2]; ok {
		t.Error("expected item 2 (# another-disabled) to NOT be selected")
	}
	if _, ok := selected[3]; !ok {
		t.Error("expected item 3 (another-enabled) to be selected")
	}
}

func TestGetFilteredIndices_NoSearch(t *testing.T) {
	m := New(mockInstaller{})
	m.itemNames = []string{"pkg-a", "pkg-b", "pkg-c"}

	indices := m.getFilteredIndices()
	if len(indices) != 3 {
		t.Fatalf("expected 3 indices, got %d", len(indices))
	}
	for i, idx := range indices {
		if idx != i {
			t.Errorf("index %d: expected %d, got %d", i, i, idx)
		}
	}
}

func TestGetFilteredIndices_WithSearch(t *testing.T) {
	m := New(mockInstaller{})
	m.itemNames = []string{"firefox", "vim", "firefox-developer", "neovim"}

	m.searchQuery = "firefox"
	indices := m.getFilteredIndices()
	if len(indices) != 2 {
		t.Fatalf("expected 2 matches for 'firefox', got %d: %v", len(indices), indices)
	}
	if indices[0] != 0 || indices[1] != 2 {
		t.Errorf("expected indices [0, 2], got %v", indices)
	}

	m.searchQuery = "vim"
	indices = m.getFilteredIndices()
	if len(indices) != 2 {
		t.Fatalf("expected 2 matches for 'vim', got %d: %v", len(indices), indices)
	}

	m.searchQuery = "nonexistent"
	indices = m.getFilteredIndices()
	if len(indices) != 0 {
		t.Errorf("expected 0 matches, got %d: %v", len(indices), indices)
	}
}

func TestHandleSelectAll(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems
	m.itemNames = []string{"pkg1", "pkg2", "pkg3"}
	// Deselect all first
	m.selectedItems = make(map[int]struct{})

	m = m.handleSelectAll()
	if len(m.selectedItems) != 3 {
		t.Errorf("expected 3 selected items, got %d", len(m.selectedItems))
	}
}

func TestHandleDeselectAll(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems
	m.itemNames = []string{"pkg1", "pkg2", "pkg3"}
	m.selectedItems = map[int]struct{}{0: {}, 1: {}, 2: {}}

	m = m.handleDeselectAll()
	if len(m.selectedItems) != 0 {
		t.Errorf("expected 0 selected items, got %d", len(m.selectedItems))
	}
}

func TestHandleInstall_NoSelection(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems
	m.selectedItems = make(map[int]struct{})

	m, _ = m.handleInstall()
	if m.currentStage != stageItems {
		t.Errorf("expected stageItems (%d), got %d — should stay on items when nothing selected", stageItems, m.currentStage)
	}
	if !m.logsVisible {
		t.Error("expected logsVisible to be true with info message")
	}
}

func TestHandleInstall_WithSelection(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems
	m.itemNames = []string{"pkg1", "pkg2", "pkg3"}
	m.selectedItems = map[int]struct{}{0: {}, 2: {}}

	m, _ = m.handleInstall()
	if m.currentStage != stageConfirm {
		t.Errorf("expected stageConfirm (%d), got %d", stageConfirm, m.currentStage)
	}
	if !m.logsVisible {
		t.Error("expected logsVisible to be true")
	}
}

func TestHandleConfirmNo(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageConfirm
	m.searchMode = true
	m.searchQuery = "test"

	m = m.handleConfirmNo()
	if m.currentStage != stageItems {
		t.Errorf("expected stageItems (%d), got %d", stageItems, m.currentStage)
	}
	if m.searchMode {
		t.Error("expected searchMode to be false")
	}
	if m.searchQuery != "" {
		t.Errorf("expected empty searchQuery, got %q", m.searchQuery)
	}
	if m.logsVisible {
		t.Error("expected logsVisible to be false")
	}
}

func TestHandleConfirmYes(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageConfirm
	m.directory = config.PkgsDir()
	m.itemNames = []string{"pkg1", "pkg2"}
	m.selectedItems = map[int]struct{}{0: {}, 1: {}}

	m, cmd := m.handleConfirmYes()
	if m.currentStage != stageInstalling {
		t.Errorf("expected stageInstalling (%d), got %d", stageInstalling, m.currentStage)
	}
	if cmd == nil {
		t.Error("expected a command from handleConfirmYes, got nil")
	}
}

func TestHandleSearchInput(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageItems

	// Simulate typing "vim"
	keyRunes := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("vim")}
	m = m.handleSearchInput(keyRunes)
	if m.searchQuery != "vim" {
		t.Errorf("expected searchQuery 'vim', got %q", m.searchQuery)
	}

	// Backspace once
	keyBackspace := tea.KeyMsg{Type: tea.KeyBackspace}
	m = m.handleSearchInput(keyBackspace)
	if m.searchQuery != "vi" {
		t.Errorf("expected searchQuery 'vi', got %q", m.searchQuery)
	}

	// Escape clears
	keyEscape := tea.KeyMsg{Type: tea.KeyEscape}
	m = m.handleSearchInput(keyEscape)
	if m.searchQuery != "" {
		t.Errorf("expected empty searchQuery, got %q", m.searchQuery)
	}
	if m.searchMode {
		t.Error("expected searchMode false after escape")
	}
}

func TestHandleCategoryEnter_Packages(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageCategory
	m.directory = config.PkgsDir()
	m.categories = []config.Category{
		{
			Name: "Test Pkgs",
			Key:  "test",
			Items: []config.Item{
				{Name: "pkg1"},
				{Name: "# pkg2"},
			},
		},
	}
	m.cursor = 0

	m, _ = m.handleCategoryEnter()

	if m.currentStage != stageItems {
		t.Errorf("expected stageItems (%d), got %d", stageItems, m.currentStage)
	}
	if len(m.itemNames) != 2 {
		t.Errorf("expected 2 item names, got %d", len(m.itemNames))
	}
	// pkg1 should be selected (not disabled), pkg2 should not
	if _, ok := m.selectedItems[0]; !ok {
		t.Error("expected item 0 (pkg1) to be selected")
	}
	if _, ok := m.selectedItems[1]; ok {
		t.Error("expected item 1 (# pkg2) to NOT be selected")
	}
}

func TestHandleCategoryEnter_Extensions(t *testing.T) {
	m := New(mockInstaller{})
	m.currentStage = stageCategory
	m.directory = config.ExtDir()
	m.categories = []config.Category{
		{
			Name: "Test Exts",
			Key:  "test",
			Items: []config.Item{
				{Name: "ext1"},
				{Name: "ext2"},
			},
		},
	}
	m.cursor = 0

	m, _ = m.handleCategoryEnter()

	if m.currentStage != stageItems {
		t.Errorf("expected stageItems (%d), got %d", stageItems, m.currentStage)
	}
	if len(m.itemNames) != 2 {
		t.Errorf("expected 2 item names, got %d", len(m.itemNames))
	}
}

func TestCursorNavigation_Menu(t *testing.T) {
	m := New(mockInstaller{})
	// Down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after one down, got %d", m.cursor)
	}
	// Up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after up, got %d", m.cursor)
	}
}

func TestBackNavigation(t *testing.T) {
	m := New(mockInstaller{})

	// Start at menu, go to category
	m.currentStage = stageCategory
	m.categoryNames = []string{"cat1", "cat2"}

	// Press back
	backMsg := tea.KeyMsg{Type: tea.KeyEscape}
	updated, _ := m.Update(backMsg)
	m = updated.(Model)
	if m.currentStage != stageMenu {
		t.Errorf("expected stageMenu (%d) after back, got %d", stageMenu, m.currentStage)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after back, got %d", m.cursor)
	}
}

func TestQuit(t *testing.T) {
	m := New(mockInstaller{})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("expected a quit command, got nil")
	}
}

// Test styles are applied without panic
func TestStyles(t *testing.T) {
	_ = listStyle
	_ = listItemSelectedStyle
	_ = installedItemStyle
	_ = searchPromptStyle
	_ = noMatchStyle
	_ = lipgloss.NewStyle
}
