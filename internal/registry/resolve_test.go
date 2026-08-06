package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func setupLocalCatalog(t *testing.T, content string) {
	t.Helper()
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestResolve_FoundLocally_NoNetwork(t *testing.T) {
	setupLocalCatalog(t, `{"templates": [{"name": "raylib-app", "description": "jogo raylib", "source": "./examples/raylib-app"}]}`)

	// Point the GitHub source at a path that doesn't exist — if Resolve found
	// the entry locally, it must never reach this and error out.
	t.Setenv(RegistryURLEnv, filepath.Join(t.TempDir(), "does-not-exist.json"))

	entry, ok := Resolve("raylib-app", false)
	if !ok {
		t.Fatal("Resolve(\"raylib-app\") = not found, want found locally")
	}
	if entry.Source != "./examples/raylib-app" {
		t.Errorf("Resolve(\"raylib-app\").Source = %q, want %q", entry.Source, "./examples/raylib-app")
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	setupLocalCatalog(t, `{"templates": [{"name": "Raylib-App", "source": "owner/repo"}]}`)
	t.Setenv(RegistryURLEnv, filepath.Join(t.TempDir(), "does-not-exist.json"))

	if _, ok := Resolve("RAYLIB-APP", false); !ok {
		t.Error("Resolve() should match names case-insensitively")
	}
}

func TestResolve_NotFound(t *testing.T) {
	setupLocalCatalog(t, `{"templates": []}`)
	t.Setenv(RegistryURLEnv, filepath.Join(t.TempDir(), "does-not-exist.json"))

	if _, ok := Resolve("ghost-template", false); ok {
		t.Error("Resolve(\"ghost-template\") = found, want not found")
	}
}

func TestLoad_MergesLocalWithEmbedded(t *testing.T) {
	setupLocalCatalog(t, `{"templates": [{"name": "local-only", "description": "só local", "source": "./x"}]}`)
	t.Setenv(RegistryURLEnv, filepath.Join(t.TempDir(), "does-not-exist.json"))

	catalog := Load(false)

	names := make(map[string]bool, len(catalog.Entries))
	for _, e := range catalog.Entries {
		names[e.Name] = true
	}
	if !names["local-only"] {
		t.Errorf("Load().Entries = %+v, want it to include the local-only entry", catalog.Entries)
	}
	if !names["raylib-app"] {
		t.Errorf("Load().Entries = %+v, want it to still include the embedded raylib-app entry", catalog.Entries)
	}
}

func TestLoad_LocalOverridesEmbeddedByName(t *testing.T) {
	setupLocalCatalog(t, `{"templates": [{"name": "raylib-app", "description": "override local", "source": "./meu-fork"}]}`)
	t.Setenv(RegistryURLEnv, filepath.Join(t.TempDir(), "does-not-exist.json"))

	catalog := Load(false)

	var found *Entry
	for i, e := range catalog.Entries {
		if e.Name == "raylib-app" {
			found = &catalog.Entries[i]
		}
	}
	if found == nil {
		t.Fatalf("Load().Entries = %+v, want a raylib-app entry", catalog.Entries)
	}
	if found.Source != "./meu-fork" {
		t.Errorf("raylib-app entry Source = %q, want the local override %q", found.Source, "./meu-fork")
	}
}
