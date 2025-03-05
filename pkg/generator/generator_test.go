package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSite(t *testing.T) {
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "templates")
	outputDir := filepath.Join(tempDir, "output")

	// Create mock template files
	mockTemplates := map[string]string{
		"components/header.tmpl":         "<header>{{.Title}}</header>",
		"components/widgets/widget.tmpl": "<widget>Widget</widget>",
		"components/footer.tmpl":         "<footer>Footer</footer>",
		"layouts/base.tmpl":              "{{template \"header.tmpl\" .}}<main>{{template %CONTENT% .}}</main>{{template \"footer.tmpl\" .}}",
		"pages/index.tmpl":               "{{- /* layout:base.tmpl */ -}}<p>Index Page</p>",
		"pages/about.tmpl":               "{{- /* layout:base.tmpl */ -}}<p>About Page</p>",
		"pages/robots.txt":               "User-agent: *\nDisallow: /",
	}

	mockData := map[string]map[string]interface{}{
		"pages/index.tmpl.json": {"Title": "Home Page"},
		"pages/about.tmpl.json": {"Title": "About Me"},
	}

	for path, content := range mockTemplates {
		fullPath := filepath.Join(templateDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), os.ModePerm); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	for path, data := range mockData {
		fullPath := filepath.Join(templateDir, path)
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		if err := os.WriteFile(fullPath, jsonData, os.ModePerm); err != nil {
			t.Fatalf("Failed to create JSON data file: %v", err)
		}
	}

	cfg := SiteConfig{
		TemplateDir: templateDir,
		OutputDir:   outputDir,
	}

	gen := NewSiteGenerator(cfg)

	if err := gen.GenerateSite(); err != nil {
		t.Fatalf("GenerateSite failed: %v", err)
	}

	expectedFiles := map[string]string{
		"index.html": "<header>Home Page</header><main><p>Index Page</p></main><footer>Footer</footer>",
		"about.html": "<header>About Me</header><main><p>About Page</p></main><footer>Footer</footer>",
		"robots.txt": "User-agent: *\nDisallow: /",
	}

	for file, expectedContent := range expectedFiles {
		outputPath := filepath.Join(outputDir, file)
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Errorf("Expected file %s not found", file)
			continue
		}
		if strings.TrimSpace(string(data)) != expectedContent {
			t.Errorf("File %s content mismatch\nExpected:\n%s\nGot:\n%s", file, expectedContent, strings.TrimSpace(string(data)))
		}
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "robots.txt")
	dstFile := filepath.Join(tempDir, "copy.txt")

	content := "User-agent: *\nDisallow: /"
	if err := os.WriteFile(srcFile, []byte(content), os.ModePerm); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s not found", dstFile)
	}
}
