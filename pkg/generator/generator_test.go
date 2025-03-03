package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSite(t *testing.T) {
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "templates")
	outputDir := filepath.Join(tempDir, "output")

	// Create mock template files
	mockTemplates := map[string]string{
		"components/header.tmpl": "<header>Header</header>",
		"components/footer.tmpl": "<footer>Footer</footer>",
		"pages/index.tmpl":       "{{template \"header.tmpl\"}}<main>Index Page</main>{{template \"footer.tmpl\"}}",
		"pages/about.tmpl":       "{{template \"header.tmpl\"}}<main>About Page</main>{{template \"footer.tmpl\"}}",
		"pages/robots.txt":       "User-agent: *\nDisallow: /",
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

	cfg := SiteConfig{
		TemplateDir: templateDir,
		OutputDir:   outputDir,
	}

	gen := NewSiteGenerator(cfg)

	if err := gen.GenerateSite(); err != nil {
		t.Fatalf("GenerateSite failed: %v", err)
	}

	expectedFiles := []string{
		"index.html",
		"about.html",
		"robots.txt",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(outputDir, file)); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found", file)
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
