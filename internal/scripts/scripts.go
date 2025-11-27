package scripts

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

func InstallVSCodeExtension(extension string) (bool, string) {
	cmd := exec.Command("codium", "--install-extension", extension)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", extension, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("%s: Installed successfully", extension)
}

func InstallPackage(pkg string) (bool, string) {
	cmd := exec.Command("paru", "-S", "--needed", "--noconfirm", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", pkg, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("%s: Installed successfully", pkg)
}
