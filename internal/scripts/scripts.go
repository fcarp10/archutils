package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	extCacheOnce sync.Once
	extCache     map[string]bool
)

// editorBinary returns the editor binary to use for extension management.
// Defaults to "codium"; override with the ARCHUTILS_EDITOR environment variable
// (e.g., "code", "code-oss", "codium").
func editorBinary() string {
	if bin := os.Getenv("ARCHUTILS_EDITOR"); bin != "" {
		return bin
	}
	return "codium"
}

// getInstalledExtensions returns a lazily-loaded set of installed VSCode/VSCodium extensions.
// The result is cached via sync.Once for thread-safe one-time initialization.
func getInstalledExtensions() map[string]bool {
	extCacheOnce.Do(func() {
		extCache = make(map[string]bool)
		cmd := exec.Command(editorBinary(), "--list-extensions")
		output, err := cmd.Output()
		if err != nil {
			return
		}
		for _, line := range strings.Split(string(output), "\n") {
			if ext := strings.TrimSpace(line); ext != "" {
				extCache[ext] = true
			}
		}
	})
	return extCache
}

// enableService runs systemctl enable --now for the given service.
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

// disableSSHPasswordAuth writes a drop-in config disabling SSH password auth.
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
