package scripts

import (
	"os/exec"
	"strings"
)

type InstalledPackage struct {
	Name        string
	Description string
}

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

type Runner struct{}

func (r Runner) InstallPackage(pkg string) (bool, string) {
	return InstallPackage(pkg)
}

func (r Runner) InstallParu() (bool, string) {
	return InstallParu()
}

func (r Runner) InstallVSCodeExtension(extension string) (bool, string) {
	return InstallVSCodeExtension(extension)
}

func (r Runner) EnableAutologin() (bool, string) {
	return EnableAutologin()
}

func (r Runner) EnablePasswordlessSSH() (bool, string) {
	return EnablePasswordlessSSH()
}

func (r Runner) EnablePasswordlessSudo() (bool, string) {
	return EnablePasswordlessSudo()
}

func (r Runner) GetPackageDescription(item string) string {
	return GetPackageDescription(item)
}

func (r Runner) GetExtensionDescription(extension string) string {
	return GetExtensionDescription(extension)
}

func (r Runner) CheckParuInstalled() (bool, string) {
	return CheckParuInstalled()
}

func (r Runner) IsPackageInstalled(pkg string) bool {
	return IsPackageInstalled(pkg)
}

func (r Runner) IsExtensionInstalled(extension string) bool {
	return IsExtensionInstalled(extension)
}

func (r Runner) SudoValidateCmd() *exec.Cmd {
	return SudoValidateCmd()
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
