package scripts

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func InstallParu() (bool, string) {
	// First, check if paru is already installed
	cmd := exec.Command("which", "paru")
	err := cmd.Run()
	if err == nil {
		return true, "Paru is already installed"
	}

	// If paru is not installed, first install dependencies
	baseDevCmd := exec.Command("sudo", "pacman", "-S", "--needed", "--noconfirm", "base-devel", "git")
	if err := baseDevCmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to install base-devel and git: %v", err)
	}

	// Clone paru from AUR
	cloneCmd := exec.Command("git", "clone", "https://aur.archlinux.org/paru.git", "/tmp/paru")
	if err := cloneCmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to clone paru repository: %v", err)
	}

	// Change directory and build paru
	buildCmd := exec.Command("sh", "-c", "cd /tmp/paru && makepkg -si --noconfirm")
	var stdout, stderr bytes.Buffer
	buildCmd.Stdout = &stdout
	buildCmd.Stderr = &stderr

	err = buildCmd.Run()
	if err != nil {
		return false, fmt.Sprintf("Installation Error:\n%v\n\nStdout:\n%s\n\nStderr:\n%s\n", err, stdout.String(), stderr.String())
	}

	// Clean up temporary directory
	os.RemoveAll("/tmp/paru")

	return true, "Paru installed successfully!"
}
