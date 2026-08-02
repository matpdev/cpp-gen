package templatescaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_CopiesFilesVerbatim(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "my-template")

	if err := Generate(dest); err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}

	cmakeContent, err := os.ReadFile(filepath.Join(dest, "CMakeLists.txt"))
	if err != nil {
		t.Fatalf("read CMakeLists.txt: %v", err)
	}
	if !strings.Contains(string(cmakeContent), "{{.Name}}") {
		t.Error("CMakeLists.txt should contain the literal placeholder {{.Name}}, got it substituted or missing")
	}

	srcPath := filepath.Join(dest, "src", "{{.NameSnake}}.cpp")
	if _, err := os.Stat(srcPath); err != nil {
		t.Errorf("expected src/{{.NameSnake}}.cpp to exist literally, stat error: %v", err)
	}

	vcpkgPath := filepath.Join(dest, "{{if .UseVCPKG}}vcpkg.json{{end}}")
	if _, err := os.Stat(vcpkgPath); err != nil {
		t.Errorf("expected the conditional vcpkg.json filename to exist literally, stat error: %v", err)
	}
}

func TestGenerate_RefusesNonEmptyDir(t *testing.T) {
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "existing.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	if err := Generate(dest); err == nil {
		t.Error("Generate() into a non-empty dir should error, got nil")
	}
}

func TestGenerate_EmptyExistingDirIsFine(t *testing.T) {
	dest := t.TempDir()

	if err := Generate(dest); err != nil {
		t.Fatalf("Generate() into an empty existing dir should succeed, got: %v", err)
	}
}
