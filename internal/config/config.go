package config

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// fsys is the filesystem interface used to read embedded and test configs.
// Both embed.FS and fstest.MapFS satisfy this interface.
type fsys interface {
	Open(name string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

var configFS fsys

var configDir = "configs"
var pkgsDir = configDir + "/packages"
var extDir = configDir + "/vscode"

// Init sets the filesystem used for reading configuration files.
// In production this is an embed.FS; in tests it can be a fstest.MapFS.
func Init(f embed.FS) {
	configFS = f
}

func PkgsDir() string {
	return pkgsDir
}

func ExtDir() string {
	return extDir
}

func ConfigDir() string {
	return configDir
}

func ReadFile(name string) ([]byte, error) {
	f, err := configFS.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// readCategoryFile opens a file once and returns the category name (from ### header)
// and the filtered lines (skipping empty lines and ## comments).
func readCategoryFile(filePath string) (categoryName string, items []string, err error) {
	file, err := configFS.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "###") {
			categoryName = strings.TrimSpace(strings.TrimPrefix(text, "###"))
			continue
		}
		if strings.HasPrefix(text, "##") {
			continue
		}
		items = append(items, text)
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return categoryName, items, nil
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
	subFiles, err := configFS.ReadDir(dir)
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
		categoryName, itemNames, err := readCategoryFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}
		category := Category{
			Name: categoryName,
			Key:  strings.TrimSuffix(subFile.Name(), ".txt"),
		}
		for _, name := range itemNames {
			category.Items = append(category.Items, Item{Name: name})
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func CategoryNames(categories []Category) []string {
	names := make([]string, len(categories))
	for i, cat := range categories {
		names[i] = cat.Name
	}
	return names
}
