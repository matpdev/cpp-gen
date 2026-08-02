package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLocal_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	entries, err := loadLocal()
	if err != nil {
		t.Fatalf("loadLocal() unexpected error for missing file: %v", err)
	}
	if entries != nil {
		t.Errorf("loadLocal() = %+v, want nil for missing file", entries)
	}
}

func TestLoadLocal_ExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path, err := LocalCatalogPath()
	if err != nil {
		t.Fatalf("LocalCatalogPath() error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `{"templates": [{"name": "meu-template", "description": "local", "source": "./local"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	entries, err := loadLocal()
	if err != nil {
		t.Fatalf("loadLocal() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "meu-template" {
		t.Errorf("loadLocal() = %+v, want single entry named meu-template", entries)
	}
}

func TestAddLocal_AppendsNew(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	if err := AddLocal(Entry{Name: "foo", Source: "./foo"}); err != nil {
		t.Fatalf("AddLocal() unexpected error: %v", err)
	}
	if err := AddLocal(Entry{Name: "bar", Source: "./bar"}); err != nil {
		t.Fatalf("AddLocal() unexpected error: %v", err)
	}

	entries, err := loadLocal()
	if err != nil {
		t.Fatalf("loadLocal() unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("loadLocal() = %+v, want 2 entries", entries)
	}
}

func TestAddLocal_UpsertsByName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")

	if err := AddLocal(Entry{Name: "foo", Source: "./v1", Description: "first"}); err != nil {
		t.Fatalf("AddLocal() unexpected error: %v", err)
	}
	if err := AddLocal(Entry{Name: "FOO", Source: "./v2", Description: "second"}); err != nil {
		t.Fatalf("AddLocal() unexpected error: %v", err)
	}

	entries, err := loadLocal()
	if err != nil {
		t.Fatalf("loadLocal() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Source != "./v2" {
		t.Errorf("loadLocal() = %+v, want single upserted entry with Source ./v2", entries)
	}
}
