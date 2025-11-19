package packages

import (
	"bufio"
	"embed"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type CategoryPkgs struct {
	Name          string
	Key           string
	PackagesNames []string
}

var PKGS_DIR = "configs/arch-pkgs"
var PackagesFS embed.FS

func ReadCategoriesPkgs() ([]CategoryPkgs, error) {
	var categories []CategoryPkgs
	subFiles, err := PackagesFS.ReadDir(PKGS_DIR)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %v", PKGS_DIR, err)
	}
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			continue
		}
		if filepath.Ext(subFile.Name()) != ".txt" {
			continue
		}
		filePath := filepath.Join(PKGS_DIR, subFile.Name())
		packages, err := readPackagesFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}
		categoryName, err := extractSubcategoryName(filePath)
		category := new(CategoryPkgs)
		if err == nil {
			category.Name = strings.TrimSuffix(categoryName, ".txt")
			category.PackagesNames = packages
			category.Key = strings.TrimSuffix(subFile.Name(), ".txt")
		}
		categories = append(categories, *category)
	}

	return categories, nil
}

func extractSubcategoryName(filePath string) (string, error) {
	file, err := PackagesFS.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(text, "###") {
			return strings.TrimSpace(strings.TrimPrefix(text, "###")), nil
		}
	}

	return "", fmt.Errorf("no subcategory name found")
}

func CategoryNames(categories []CategoryPkgs) []string {
	names := make([]string, len(categories))
	for i, cat := range categories {
		names[i] = cat.Name
	}
	return names
}

func readPackagesFromFile(filePath string) ([]string, error) {
	file, err := PackagesFS.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if strings.HasPrefix(text, "###") || strings.HasPrefix(text, "##") {
			continue
		}

		packages = append(packages, text)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}

func InstallPackage(pkg string, ext bool) (bool, string) {
	var cmd *exec.Cmd
	if !ext {
		cmd = exec.Command("paru", "-S", "--needed", "--noconfirm", pkg)
	} else {
		cmd = exec.Command("codium", "--install-extension", pkg)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("%s: Failed to install %v\n%s", pkg, err, strings.Trim(string(output), "\n"))
	}
	return true, fmt.Sprintf("%s: Installed successfully", pkg)
}
