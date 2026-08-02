package customtemplate

import (
	"path/filepath"
	"testing"
)

func TestParseSource_Local(t *testing.T) {
	dir := t.TempDir()

	cases := []string{
		dir,
		".",
		"./sub",
		"~",
	}

	for _, raw := range cases {
		src, err := ParseSource(raw)
		if err != nil {
			t.Fatalf("ParseSource(%q) unexpected error: %v", raw, err)
		}
		if src.Kind != SourceLocal {
			t.Errorf("ParseSource(%q).Kind = %v, want SourceLocal", raw, src.Kind)
		}
		if !filepath.IsAbs(src.LocalPath) {
			t.Errorf("ParseSource(%q).LocalPath = %q, want absolute path", raw, src.LocalPath)
		}
	}
}

func TestParseSource_GitHubShorthand(t *testing.T) {
	cases := []struct {
		raw        string
		wantURL    string
		wantRef    string
		wantSubdir string
	}{
		{"owner/repo", "https://github.com/owner/repo.git", "", ""},
		{"owner/repo#branch", "https://github.com/owner/repo.git", "branch", ""},
		{"owner/repo/sub#branch", "https://github.com/owner/repo.git", "branch", "sub"},
		{"owner/repo/sub/dir", "https://github.com/owner/repo.git", "", "sub/dir"},
	}

	for _, c := range cases {
		src, err := ParseSource(c.raw)
		if err != nil {
			t.Fatalf("ParseSource(%q) unexpected error: %v", c.raw, err)
		}
		if src.Kind != SourceGit {
			t.Errorf("ParseSource(%q).Kind = %v, want SourceGit", c.raw, src.Kind)
		}
		if src.GitURL != c.wantURL {
			t.Errorf("ParseSource(%q).GitURL = %q, want %q", c.raw, src.GitURL, c.wantURL)
		}
		if src.Ref != c.wantRef {
			t.Errorf("ParseSource(%q).Ref = %q, want %q", c.raw, src.Ref, c.wantRef)
		}
		if src.Subdir != c.wantSubdir {
			t.Errorf("ParseSource(%q).Subdir = %q, want %q", c.raw, src.Subdir, c.wantSubdir)
		}
	}
}

func TestParseSource_FullURLs(t *testing.T) {
	cases := []struct {
		raw     string
		wantURL string
		wantRef string
	}{
		{"https://github.com/owner/repo.git", "https://github.com/owner/repo.git", ""},
		{"https://github.com/owner/repo.git#v1.0.0", "https://github.com/owner/repo.git", "v1.0.0"},
		{"git@github.com:owner/repo.git", "git@github.com:owner/repo.git", ""},
		{"git@github.com:owner/repo.git#main", "git@github.com:owner/repo.git", "main"},
	}

	for _, c := range cases {
		src, err := ParseSource(c.raw)
		if err != nil {
			t.Fatalf("ParseSource(%q) unexpected error: %v", c.raw, err)
		}
		if src.Kind != SourceGit {
			t.Errorf("ParseSource(%q).Kind = %v, want SourceGit", c.raw, src.Kind)
		}
		if src.GitURL != c.wantURL {
			t.Errorf("ParseSource(%q).GitURL = %q, want %q", c.raw, src.GitURL, c.wantURL)
		}
		if src.Ref != c.wantRef {
			t.Errorf("ParseSource(%q).Ref = %q, want %q", c.raw, src.Ref, c.wantRef)
		}
	}
}

func TestParseSource_Invalid(t *testing.T) {
	cases := []string{"", "   ", "justaname", "#branch"}

	for _, raw := range cases {
		if _, err := ParseSource(raw); err == nil {
			t.Errorf("ParseSource(%q) expected error, got nil", raw)
		}
	}
}
