package registry

import "testing"

func TestCatalog_Search(t *testing.T) {
	c := Catalog{Entries: []Entry{
		{Name: "raylib-app", Description: "Jogo simples com raylib", Tags: []string{"game", "graphics"}},
		{Name: "http-server", Description: "Servidor HTTP com Asio", Tags: []string{"network"}},
	}}

	cases := []struct {
		query string
		want  []string
	}{
		{"", []string{"raylib-app", "http-server"}},
		{"raylib", []string{"raylib-app"}},
		{"GAME", []string{"raylib-app"}},
		{"network", []string{"http-server"}},
		{"nonexistent", nil},
	}

	for _, c2 := range cases {
		got := c.Search(c2.query)
		if len(got) != len(c2.want) {
			t.Fatalf("Search(%q) = %d entries, want %d", c2.query, len(got), len(c2.want))
		}
		for i, e := range got {
			if e.Name != c2.want[i] {
				t.Errorf("Search(%q)[%d].Name = %q, want %q", c2.query, i, e.Name, c2.want[i])
			}
		}
	}
}

func TestParseCatalogJSON(t *testing.T) {
	data := []byte(`{"templates": [{"name": "foo", "description": "bar", "source": "owner/foo", "tags": ["x"]}]}`)
	entries, err := parseCatalogJSON(data)
	if err != nil {
		t.Fatalf("parseCatalogJSON() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "foo" || entries[0].Source != "owner/foo" {
		t.Errorf("parseCatalogJSON() = %+v, want single entry named foo", entries)
	}
}

func TestParseCatalogJSON_Invalid(t *testing.T) {
	if _, err := parseCatalogJSON([]byte("not json")); err == nil {
		t.Error("parseCatalogJSON(invalid) expected error, got nil")
	}
}

func TestEmbeddedEntries_StartsEmpty(t *testing.T) {
	// Guards the packaging assumption documented in registry.go: the
	// bundled catalog.json ships empty until maintainers curate entries.
	if entries := embeddedEntries(false); len(entries) != 0 {
		t.Errorf("embeddedEntries() = %d entries, want 0 (seed catalog.json is empty by design)", len(entries))
	}
}
