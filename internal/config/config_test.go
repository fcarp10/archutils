package config

import (
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"configs/packages/01-wm.txt": {
			Data: []byte("### Window Managers\n## Hyprland\n# hyprland\nniri\n\nwaybar\nmako\n"),
		},
		"configs/packages/02-cli.txt": {
			Data: []byte("### CLI Tools\n## Git\ngit-delta\n\n## Shell\nzsh\nfish\n"),
		},
		"configs/packages/03-empty.txt": {
			Data: []byte("### Empty Category\n"),
		},
		"configs/packages/04-no-header.txt": {
			Data: []byte("pkg1\npkg2\n"),
		},
		"configs/vscode/01-languages.txt": {
			Data: []byte("### Languages\ngolang.go\nredhat.java\n"),
		},
	}
}

func TestReadCategoryFile(t *testing.T) {
	fs := testFS()
	configFS = fs

	tests := []struct {
		filePath  string
		wantName  string
		wantItems []string
	}{
		{
			filePath:  "configs/packages/01-wm.txt",
			wantName:  "Window Managers",
			wantItems: []string{"# hyprland", "niri", "waybar", "mako"},
		},
		{
			filePath:  "configs/packages/02-cli.txt",
			wantName:  "CLI Tools",
			wantItems: []string{"git-delta", "zsh", "fish"},
		},
		{
			filePath:  "configs/packages/03-empty.txt",
			wantName:  "Empty Category",
			wantItems: nil,
		},
		{
			filePath:  "configs/packages/04-no-header.txt",
			wantName:  "",
			wantItems: []string{"pkg1", "pkg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			name, items, err := readCategoryFile(tt.filePath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if name != tt.wantName {
				t.Errorf("name: expected %q, got %q", tt.wantName, name)
			}
			if len(items) != len(tt.wantItems) {
				t.Fatalf("items: expected %d, got %d: %v", len(tt.wantItems), len(items), items)
			}
			for i, want := range tt.wantItems {
				if items[i] != want {
					t.Errorf("item %d: expected %q, got %q", i, want, items[i])
				}
			}
		})
	}
}

func TestReadCategoryFile_NotFound(t *testing.T) {
	fs := testFS()
	configFS = fs

	_, _, err := readCategoryFile("configs/packages/nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestReadCategories(t *testing.T) {
	fs := testFS()
	configFS = fs

	categories, err := ReadCategories("configs/packages")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(categories) != 4 {
		t.Fatalf("expected 4 categories, got %d", len(categories))
	}

	// Category 1: Window Managers — items are filtered (no ## lines)
	if categories[0].Name != "Window Managers" {
		t.Errorf("expected name 'Window Managers', got %q", categories[0].Name)
	}
	if categories[0].Key != "01-wm" {
		t.Errorf("expected key '01-wm', got %q", categories[0].Key)
	}
	if len(categories[0].Items) != 4 {
		t.Errorf("expected 4 items, got %d: %v", len(categories[0].Items), categories[0].Items)
	}

	// Category 2: CLI Tools
	if categories[1].Name != "CLI Tools" {
		t.Errorf("expected name 'CLI Tools', got %q", categories[1].Name)
	}
	if len(categories[1].Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(categories[1].Items))
	}

	// Category 3: Empty Category
	if categories[2].Name != "Empty Category" {
		t.Errorf("expected name 'Empty Category', got %q", categories[2].Name)
	}
	if len(categories[2].Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(categories[2].Items))
	}

	// Category 4: No header — name is empty but items are still read in single pass
	if categories[3].Name != "" {
		t.Errorf("expected empty name, got %q", categories[3].Name)
	}
	if len(categories[3].Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(categories[3].Items))
	}
}

func TestReadCategories_EmptyDir(t *testing.T) {
	fs := fstest.MapFS{
		"configs/packages/.gitkeep": {Data: []byte{}},
	}
	configFS = fs

	categories, err := ReadCategories("configs/packages")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(categories) != 0 {
		t.Errorf("expected 0 categories, got %d", len(categories))
	}
}

func TestCategoryNames(t *testing.T) {
	categories := []Category{
		{Name: "Window Managers", Key: "01-wm"},
		{Name: "CLI Tools", Key: "02-cli"},
		{Name: "Browsers", Key: "03-browsers"},
	}

	names := CategoryNames(categories)
	expected := []string{"Window Managers", "CLI Tools", "Browsers"}

	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("index %d: expected %q, got %q", i, want, names[i])
		}
	}
}

func TestCategoryNames_Empty(t *testing.T) {
	names := CategoryNames(nil)
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}

	names = CategoryNames([]Category{})
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

func TestReadFile(t *testing.T) {
	fs := testFS()
	configFS = fs

	data, err := ReadFile("configs/packages/01-wm.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}
}

func TestReadFile_NotFound(t *testing.T) {
	fs := testFS()
	configFS = fs

	_, err := ReadFile("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestDirPaths(t *testing.T) {
	if PkgsDir() != "configs/packages" {
		t.Errorf("expected 'configs/packages', got %q", PkgsDir())
	}
	if ExtDir() != "configs/vscode" {
		t.Errorf("expected 'configs/vscode', got %q", ExtDir())
	}
	if ConfigDir() != "configs" {
		t.Errorf("expected 'configs', got %q", ConfigDir())
	}
}
