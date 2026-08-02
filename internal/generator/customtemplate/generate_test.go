package customtemplate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_ConditionalFileSkip(t *testing.T) {
	templateRoot := t.TempDir()

	mustWrite(t, filepath.Join(templateRoot, "{{if .UseVCPKG}}vcpkg.json{{end}}"), `{"name": "{{.Name}}"}`)
	mustWrite(t, filepath.Join(templateRoot, "{{if .UseFetchContent}}cmake-fc-only.txt{{end}}"), "fetchcontent stuff")
	mustWrite(t, filepath.Join(templateRoot, "always.txt"), "Hello {{.Name}}")

	destRoot := t.TempDir()
	data := struct {
		Name            string
		UseVCPKG        bool
		UseFetchContent bool
	}{Name: "demo", UseVCPKG: true, UseFetchContent: false}

	if err := Generate(templateRoot, destRoot, data, false); err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}

	assertFileContains(t, filepath.Join(destRoot, "vcpkg.json"), `"name": "demo"`)
	assertFileContains(t, filepath.Join(destRoot, "always.txt"), "Hello demo")

	if _, err := os.Stat(filepath.Join(destRoot, "cmake-fc-only.txt")); !os.IsNotExist(err) {
		t.Errorf("cmake-fc-only.txt should have been skipped (UseFetchContent=false), stat err = %v", err)
	}
}

func TestGenerate_ConditionalDirSkip(t *testing.T) {
	templateRoot := t.TempDir()

	mustWrite(t, filepath.Join(templateRoot, "{{if .UseFetchContent}}cmake{{end}}", "Dependencies.cmake"), "fetchcontent deps")
	mustWrite(t, filepath.Join(templateRoot, "always.txt"), "kept")

	destRoot := t.TempDir()
	data := struct{ UseFetchContent bool }{UseFetchContent: false}

	if err := Generate(templateRoot, destRoot, data, false); err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destRoot, "cmake")); !os.IsNotExist(err) {
		t.Errorf("cmake/ dir should have been skipped entirely, stat err = %v", err)
	}
	assertFileContains(t, filepath.Join(destRoot, "always.txt"), "kept")
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(got), want) {
		t.Errorf("%s content = %q, want it to contain %q", path, string(got), want)
	}
}
