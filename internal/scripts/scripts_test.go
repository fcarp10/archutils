package scripts

import (
	"os"
	"testing"
)

func TestEditorBinary_Default(t *testing.T) {
	// Ensure env var is unset
	os.Unsetenv("ARCHUTILS_EDITOR")
	got := editorBinary()
	if got != "codium" {
		t.Errorf("expected default 'codium', got %q", got)
	}
}

func TestEditorBinary_EnvOverride(t *testing.T) {
	os.Setenv("ARCHUTILS_EDITOR", "code")
	defer os.Unsetenv("ARCHUTILS_EDITOR")

	got := editorBinary()
	if got != "code" {
		t.Errorf("expected 'code' from env, got %q", got)
	}
}

func TestEditorBinary_EnvOverrideCustom(t *testing.T) {
	os.Setenv("ARCHUTILS_EDITOR", "code-oss")
	defer os.Unsetenv("ARCHUTILS_EDITOR")

	got := editorBinary()
	if got != "code-oss" {
		t.Errorf("expected 'code-oss' from env, got %q", got)
	}
}
