package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matpdev/cpp-gen/internal/localconfig"
)

// LocalCatalogFileName is the file name for the user-editable local catalog,
// stored under localconfig.GlobalConfigDir() (e.g. ~/.config/cpp-gen/templates.json).
const LocalCatalogFileName = "templates.json"

// LocalCatalogPath returns the path to the user's local template catalog.
func LocalCatalogPath() (string, error) {
	dir, err := localconfig.GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, LocalCatalogFileName), nil
}

// loadLocal reads the user's local catalog. A missing file is not an error —
// it just means the user hasn't added any local entries yet.
func loadLocal() ([]Entry, error) {
	path, err := LocalCatalogPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return parseCatalogJSON(data)
}

// SaveLocal overwrites the user's local catalog with entries, creating the
// config directory if needed.
func SaveLocal(entries []Entry) error {
	path, err := LocalCatalogPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("criar diretório de config %s: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(catalogFile{Templates: entries}, "", "    ")
	if err != nil {
		return fmt.Errorf("serializar catálogo local: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("escrever %s: %w", path, err)
	}

	return nil
}

// AddLocal upserts entry (matched by case-insensitive name) into the user's
// local catalog, preserving every other existing entry.
func AddLocal(entry Entry) error {
	existing, err := loadLocal()
	if err != nil {
		return err
	}

	key := strings.ToLower(entry.Name)
	replaced := false
	for i, e := range existing {
		if strings.ToLower(e.Name) == key {
			existing[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		existing = append(existing, entry)
	}

	return SaveLocal(existing)
}
