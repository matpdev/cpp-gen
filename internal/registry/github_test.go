package registry

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGitHub_LocalFileOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	content := `{"templates": [{"name": "dev-override", "description": "via arquivo local", "source": "owner/repo"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	t.Setenv(RegistryURLEnv, path)

	entries, err := loadGitHub()
	if err != nil {
		t.Fatalf("loadGitHub() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "dev-override" {
		t.Errorf("loadGitHub() = %+v, want single entry named dev-override", entries)
	}
}

func TestLoadGitHub_HTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"templates": [{"name": "remote-tpl", "description": "via http", "source": "owner/repo"}]}`))
	}))
	defer srv.Close()

	t.Setenv(RegistryURLEnv, srv.URL)

	entries, err := loadGitHub()
	if err != nil {
		t.Fatalf("loadGitHub() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "remote-tpl" {
		t.Errorf("loadGitHub() = %+v, want single entry named remote-tpl", entries)
	}
}

func TestLoadGitHub_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv(RegistryURLEnv, srv.URL)

	if _, err := loadGitHub(); err == nil {
		t.Error("loadGitHub() expected error for HTTP 404, got nil")
	}
}
