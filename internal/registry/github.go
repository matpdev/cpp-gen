package registry

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// defaultRegistryURL points to the raw JSON catalog of the community
// registry. It doesn't need to exist for cpp-gen to work — a failed fetch is
// treated as "no GitHub entries available", never a hard error.
const defaultRegistryURL = "https://raw.githubusercontent.com/matpdev/cpp-gen-templates/main/templates.json"

// RegistryURLEnv overrides the registry location. Accepts an http(s) URL
// (fetched as-is) or a local file path (read directly) — the latter is
// handy for testing a registry file before publishing it, the same way
// CPPGEN_VULKAN_TEMPLATES_DIR lets you iterate on the Vulkan template
// without recompiling.
const RegistryURLEnv = "CPPGEN_REGISTRY_URL"

// fetchTimeout bounds how long a registry lookup can block `cpp-gen new`.
const fetchTimeout = 5 * time.Second

// registryURL resolves the effective registry location.
func registryURL() string {
	if v := os.Getenv(RegistryURLEnv); v != "" {
		return v
	}
	return defaultRegistryURL
}

// loadGitHub fetches and parses the registry catalog.
func loadGitHub() ([]Entry, error) {
	url := registryURL()

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		data, err := os.ReadFile(url)
		if err != nil {
			return nil, fmt.Errorf("ler registro local %q: %w", url, err)
		}
		return parseCatalogJSON(data)
	}

	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("buscar registro em %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registro em %s retornou HTTP %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ler resposta de %s: %w", url, err)
	}

	return parseCatalogJSON(data)
}
