package customtemplate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Fetch resolves a Source into a local directory ready to be read by
// Generate, plus a cleanup func that must be called once the caller is done
// with it.
//
// Local sources are returned as-is (cleanup is a no-op — a user's own folder
// is never deleted). Git sources are shallow-cloned into a temporary
// directory that cleanup removes.
func Fetch(src Source) (dir string, cleanup func(), err error) {
	switch src.Kind {
	case SourceLocal:
		info, statErr := os.Stat(src.LocalPath)
		if statErr != nil {
			return "", nil, fmt.Errorf("template local %q: %w", src.LocalPath, statErr)
		}
		if !info.IsDir() {
			return "", nil, fmt.Errorf("template local %q não é um diretório", src.LocalPath)
		}
		return src.LocalPath, func() {}, nil

	case SourceGit:
		return fetchGit(src)

	default:
		return "", nil, fmt.Errorf("tipo de fonte de template desconhecido")
	}
}

func fetchGit(src Source) (dir string, cleanup func(), err error) {
	tmpDir, mkErr := os.MkdirTemp("", "cppgen-template-*")
	if mkErr != nil {
		return "", nil, fmt.Errorf("criar diretório temporário: %w", mkErr)
	}
	cleanup = func() { os.RemoveAll(tmpDir) }

	args := []string{"clone", "--depth", "1"}
	if src.Ref != "" {
		args = append(args, "--branch", src.Ref)
	}
	args = append(args, src.GitURL, tmpDir)

	cmd := exec.Command("git", args...)
	output, cloneErr := cmd.CombinedOutput()
	if cloneErr != nil {
		cleanup()
		return "", nil, fmt.Errorf("git clone %s: %w\n%s", src.GitURL, cloneErr, string(output))
	}

	root := tmpDir
	if src.Subdir != "" {
		root = filepath.Join(tmpDir, src.Subdir)
		if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
			cleanup()
			return "", nil, fmt.Errorf("subdiretório %q não encontrado em %s", src.Subdir, src.GitURL)
		}
	}

	return root, cleanup, nil
}
