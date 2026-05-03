package scripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	c "github.com/fcarp10/archutils/internal/config"
)

type Installer interface {
	InstallPackage(pkg string) (bool, string)
	InstallParu() (bool, string)
	InstallVSCodeExtension(extension string) (bool, string)
	EnableAutologin() (bool, string)
	EnablePasswordlessSSH() (bool, string)
	EnablePasswordlessSudo() (bool, string)
	GetPackageDescription(item string) string
	GetExtensionDescription(extension string) string
	CheckParuInstalled() (bool, string)
	IsPackageInstalled(pkg string) bool
	IsExtensionInstalled(extension string) bool
	SudoValidateCmd() *exec.Cmd
	GetInstalledPackages() map[string]string
}

// Runner implements the Installer interface by executing system commands.
type Runner struct{}

func (r Runner) InstallPackage(pkg string) (bool, string) {
	if ok, msg := r.CheckParuInstalled(); !ok {
		return false, msg
	}

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
	cmd := exec.Command("paru", "-S", "--needed", "--noconfirm", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", packageName, err, strings.Trim(string(output), "\n"))
	}
	message := fmt.Sprintf("%s: Installed successfully", packageName)
	for _, service := range services {
		success, enableMsg := enableService(service, userLevel)
		if !success {
			return false, fmt.Sprintf("%s\n%s", message, enableMsg)
		}
		message += "\n" + enableMsg
	}
	return true, message
}

func (r Runner) InstallParu() (bool, string) {
	checkOutput, checkErr := exec.Command("pacman", "-Qeq", "paru").CombinedOutput()
	if checkErr == nil && len(checkOutput) > 0 {
		pkgs := strings.TrimSpace(string(checkOutput))
		cleanCmd := exec.Command("sudo", "pacman", "-Rns", "--noconfirm", pkgs)
		cleanOutput, cleanErr := cleanCmd.CombinedOutput()
		if cleanErr != nil {
			return false, fmt.Sprintf("Failed to uninstall previous versions of paru (%s): %v\n%s", pkgs, cleanErr, strings.TrimSpace(string(cleanOutput)))
		}
	}

	baseDevCmd := exec.Command("sudo", "pacman", "-S", "--needed", "--noconfirm", "base-devel", "git")
	if err := baseDevCmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to install base-devel and git: %v", err)
	}

	cloneCmd := exec.Command("git", "clone", "https://aur.archlinux.org/paru.git", "/tmp/paru")
	if err := cloneCmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to clone paru repository: %v", err)
	}

	buildCmd := exec.Command("makepkg", "-si", "--noconfirm")
	buildCmd.Dir = "/tmp/paru"
	var stdout, stderr bytes.Buffer
	buildCmd.Stdout = &stdout
	buildCmd.Stderr = &stderr

	err := buildCmd.Run()
	if err != nil {
		return false, fmt.Sprintf("Installation Error:\n%v\n\nStdout:\n%s\n\nStderr:\n%s\n", err, stdout.String(), stderr.String())
	}

	os.RemoveAll("/tmp/paru")
	return true, "Paru installed successfully!"
}

func (r Runner) InstallVSCodeExtension(extension string) (bool, string) {
	cmd := exec.Command(editorBinary(), "--install-extension", extension)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", extension, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("%s: Installed successfully", extension)
}

func (r Runner) EnableAutologin() (bool, string) {
	user := os.Getenv("USER")
	if user == "" {
		return false, "Unable to get current user"
	}

	content, err := c.ReadFile(c.ConfigDir() + "/autologin.conf")
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

func (r Runner) EnablePasswordlessSSH() (bool, string) {
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

func (r Runner) EnablePasswordlessSudo() (bool, string) {
	user := os.Getenv("USER")
	if user == "" {
		return false, "Unable to get current user"
	}

	sudoersPath := fmt.Sprintf("/etc/sudoers.d/%s", user)

	if _, err := os.Stat(sudoersPath); err == nil {
		return true, "Passwordless sudo is already configured"
	}

	sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD: ALL\n", user)

	cmd := exec.Command("sudo", "tee", sudoersPath)
	cmd.Stdin = strings.NewReader(sudoersContent)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to configure passwordless sudo: %v\n%s", err, strings.TrimSpace(stderr.String()))
	}

	checkSudoersCmd := exec.Command("sudo", "visudo", "-c", "-f", sudoersPath)
	if checkOutput, checkErr := checkSudoersCmd.CombinedOutput(); checkErr != nil {
		exec.Command("sudo", "rm", "-f", sudoersPath).Run()
		return false, fmt.Sprintf("Sudoers file syntax error: %s", strings.TrimSpace(string(checkOutput)))
	}

	permCmd := exec.Command("sudo", "chmod", "440", sudoersPath)
	if err := permCmd.Run(); err != nil {
		return false, fmt.Sprintf("Failed to set permissions on sudoers file: %v", err)
	}

	return true, "Passwordless sudo configured successfully"
}

func (r Runner) GetPackageDescription(item string) string {
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

func (r Runner) GetExtensionDescription(extension string) string {
	fields := strings.Fields(extension)
	if len(fields) == 0 {
		return ""
	}
	extID := fields[0]

	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}

	dirs := []string{
		filepath.Join(home, ".local", "share", "VSCodium", "extensions"),
		filepath.Join(home, ".vscode-oss", "extensions"),
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dirName := entry.Name()
			if strings.HasPrefix(dirName, extID+"-") {
				pkgPath := filepath.Join(dir, dirName, "package.json")
				data, err := os.ReadFile(pkgPath)
				if err != nil {
					continue
				}
				var pkg struct {
					Description string `json:"description"`
				}
				if err := json.Unmarshal(data, &pkg); err != nil {
					continue
				}
				return pkg.Description
			}
		}
	}
	return ""
}

func (r Runner) CheckParuInstalled() (bool, string) {
	_, err := exec.LookPath("paru")
	if err != nil {
		return false, "paru is not installed. Please select 'Install Paru' from the main menu first."
	}
	return true, ""
}

func (r Runner) IsPackageInstalled(pkg string) bool {
	fields := strings.Fields(pkg)
	if len(fields) == 0 {
		return false
	}
	cmd := exec.Command("pacman", "-Q", fields[0])
	return cmd.Run() == nil
}

func (r Runner) IsExtensionInstalled(extension string) bool {
	fields := strings.Fields(extension)
	if len(fields) == 0 {
		return false
	}
	return getInstalledExtensions()[fields[0]]
}

func (r Runner) SudoValidateCmd() *exec.Cmd {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}

// GetInstalledPackages runs pacman -Qi once and returns a map of
// installed package names to their descriptions.
func (r Runner) GetInstalledPackages() map[string]string {
	cmd := exec.Command("pacman", "-Qi")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	result := make(map[string]string)
	var currentName string

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentName = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Description") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && currentName != "" {
				result[currentName] = strings.TrimSpace(parts[1])
			}
		}
	}
	return result
}
