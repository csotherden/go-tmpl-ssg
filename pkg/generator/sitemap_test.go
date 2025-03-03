package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSitemap(t *testing.T) {
	tempDir := t.TempDir()
	htmlFiles := []string{
		filepath.Join(tempDir, "index.html"),
		filepath.Join(tempDir, "about.html"),
	}
	baseURL := "https://example.com"

	for _, file := range htmlFiles {
		if err := os.WriteFile(file, []byte("test"), os.ModePerm); err != nil {
			t.Fatalf("Failed to create test HTML file: %v", err)
		}
	}

	if err := generateSitemap(tempDir, baseURL, htmlFiles); err != nil {
		t.Fatalf("generateSitemap failed: %v", err)
	}

	sitemapPath := filepath.Join(tempDir, "sitemap.xml")
	data, err := os.ReadFile(sitemapPath)
	if err != nil {
		t.Fatalf("Failed to read generated sitemap.xml: %v", err)
	}

	expectedContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset>
  <url>
    <loc>https://example.com/index.html</loc>
  </url>
  <url>
    <loc>https://example.com/about.html</loc>
  </url>
</urlset>`

	cleanData := strings.TrimSpace(string(data))
	expectedClean := strings.TrimSpace(expectedContent)

	if !bytes.Equal([]byte(cleanData), []byte(expectedClean)) {
		t.Errorf("Generated sitemap.xml does not match expected output\nExpected:\n%s\nGot:\n%s", expectedClean, cleanData)
	}
}
