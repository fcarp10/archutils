package scripts

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	c "github.com/fcarp10/archutils/internal/config"
)

func InstallParu() (bool, string) {

	// Check if another version of paru is installed and remove it
	checkOutput, checkErr := exec.Command("pacman", "-Qeq", "paru").CombinedOutput()
	if checkErr == nil && len(checkOutput) > 0 {
		pkgs := strings.TrimSpace(string(checkOutput))
		cleanCmd := exec.Command("sudo", "pacman", "-Rns", "--noconfirm", pkgs)
		cleanOutput, cleanErr := cleanCmd.CombinedOutput()
		if cleanErr != nil {
			return false, fmt.Sprintf("Failed to uninstall previous versions of paru (%s): %v\n%s", pkgs, cleanErr, strings.TrimSpace(string(cleanOutput)))
		}
	}

	// Install dependencies
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

	err := buildCmd.Run()
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
		return false, fmt.Sprintf("Failed to enable \033[31m%s\033[0m: %v\n%s", service, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("\033[32m%s\033[0m Enabled successfully", service)
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

func EnableAutologin() (bool, string) {
	user := os.Getenv("USER")
	if user == "" {
		return false, "Unable to get current user"
	}

	content, err := c.ConfigFS.ReadFile(c.CONFIG_DIR + "/autologin.conf")
	if err != nil {
		return false, fmt.Sprintf("Failed to read autologin.conf: %v", err)
	}

	template := string(content)
	replaced := strings.ReplaceAll(template, "$USER", user)

	cmd1 := exec.Command("sudo", "mkdir", "-p", "/etc/systemd/system/getty@tty1.service.d")
	if err := cmd1.Run(); err != nil {
		return false, fmt.Sprintf("Failed to create directory: %v", err)
	}

	cmd2 := exec.Command("sudo", "tee", "/etc/systemd/system/getty@tty1.service.d/autologin.conf")
	cmd2.Stdin = strings.NewReader(replaced)
	if err := cmd2.Run(); err != nil {
		return false, fmt.Sprintf("Failed to write autologin.conf: %v", err)
	}

	return true, "Autologin configured successfully"
}

func disableSSHPasswordAuth() (bool, string) {
	cmd1 := exec.Command("sudo", "mkdir", "-p", "/etc/ssh/ssh_config.d")
	if err := cmd1.Run(); err != nil {
		return false, fmt.Sprintf("Failed to create directory /etc/ssh/ssh_config.d: %v", err)
	}

	configContent := "PasswordAuthentication no\n"
	cmd2 := exec.Command("sudo", "tee", "/etc/ssh/ssh_config.d/disable_password.conf")
	cmd2.Stdin = strings.NewReader(configContent)
	if err := cmd2.Run(); err != nil {
		return false, fmt.Sprintf("Failed to write disable_password.conf: %v", err)
	}

	return true, "SSH password authentication disabled successfully"
}

func EnablePasswordlessSSH() (bool, string) {
	success1, msg1 := disableSSHPasswordAuth()
	if !success1 {
		return false, msg1
	}
	success2, msg2 := enableService("sshd", false)
	if !success2 {
		return false, msg1 + " - " + msg2
	}
	return true, msg1 + " - " + msg2
}
