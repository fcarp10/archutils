package config

import (
	"bufio"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

var ConfigFS embed.FS

var CONFIG_DIR = "configs"
var PKGS_DIR = CONFIG_DIR + "/packages"
var EXT_DIR = CONFIG_DIR + "/vscode"

func ReadFilteredLines(filePath string) ([]string, error) {
	file, err := ConfigFS.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "##") {
			continue
		}
		lines = append(lines, text)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

type Item struct {
	Name        string
	Description string
}

type Category struct {
	Name  string
	Key   string
	Items []Item
}

func ReadCategories(dir string) ([]Category, error) {
	var categories []Category
	subFiles, err := ConfigFS.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %v", dir, err)
	}
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			continue
		}
		if filepath.Ext(subFile.Name()) != ".txt" {
			continue
		}
		filePath := filepath.Join(dir, subFile.Name())
		items, err := ReadFilteredLines(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}
		categoryName, err := extractSubcategoryName(filePath)
		category := new(Category)
		if err == nil {
			category.Name = strings.TrimSuffix(categoryName, ".txt")
			var itemList []Item
			for _, itemName := range items {
				item := Item{Name: itemName}
				itemList = append(itemList, item)
			}
			category.Items = itemList
			category.Key = strings.TrimSuffix(subFile.Name(), ".txt")
		}
		categories = append(categories, *category)
	}

	return categories, nil
}

func extractSubcategoryName(filePath string) (string, error) {
	file, err := ConfigFS.Open(filePath)
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

func CategoryNames(categories []Category) []string {
	names := make([]string, len(categories))
	for i, cat := range categories {
		names[i] = cat.Name
	}
	return names
}
