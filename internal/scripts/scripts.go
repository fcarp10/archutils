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

func enableService(service string, userLevel bool) (bool, string) {
	var cmd *exec.Cmd
	if userLevel {
		cmd = exec.Command("systemctl", "--user", "enable", "--now", service)
	} else {
		cmd = exec.Command("sudo", "systemctl", "enable", "--now", service)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Failed to enable %s: %v\n%s", service, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("%s: Enabled successfully", service)
}

func InstallPackage(pkg string) (bool, string) {
	fields := strings.Fields(pkg)
	if len(fields) == 0 {
		return false, "Invalid package string"
	}
	packageName := fields[0]
	services := []string{}
	userLevel := false
	for i := 1; i < len(fields); i++ {
		field := fields[i]
		if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
			content := strings.TrimSuffix(strings.TrimPrefix(field, "["), "]")
			if content == "user" {
				userLevel = true
			} else {
				services = append(services, content)
			}
		}
	}
	// Install package
	cmd := exec.Command("paru", "-S", "--needed", "--noconfirm", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", packageName, err, strings.Trim(string(output), "\n"))
	}
	message := fmt.Sprintf("%s: Installed successfully", packageName)
	// Enable services if any
	for _, service := range services {
		success, enableMsg := enableService(service, userLevel)
		if !success {
			return false, fmt.Sprintf("%s\n%s", message, enableMsg)
		}
		message += "\n" + enableMsg
	}
	return true, message
}

func GetPackageDescription(item string) string {
	cmd := exec.Command("pacman", "-Q", "--info", item)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Description") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
