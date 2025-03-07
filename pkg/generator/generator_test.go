package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSite(t *testing.T) {
	type test struct {
		name          string
		templates     map[string]string
		data          map[string]interface{}
		expectedFiles map[string]string
		expectedErr   error
	}

	tests := []test{
		{
			name: "Mono test",
			templates: map[string]string{
				"components/header.tmpl":         "<header>{{.Title}}</header>",
				"components/widgets/widget.tmpl": "<widget>Widget</widget>",
				"components/footer.tmpl":         "<footer>Footer</footer>",
				"layouts/base.tmpl":              "{{template \"header.tmpl\" .}}<main>{{template %CONTENT% .}}</main>{{template \"footer.tmpl\" .}}",
				"pages/index.tmpl":               "{{- /* layout:base.tmpl */ -}}<p>Index Page</p>",
				"pages/about.tmpl":               "{{- /* layout:base.tmpl */ -}}<p>About Page</p>",
			},
			data: map[string]interface{}{
				"pages/index.tmpl.json": map[string]interface{}{"Title": "Home Page"},
				"pages/about.tmpl.json": map[string]interface{}{"Title": "About Me"},
			},
			expectedFiles: map[string]string{
				"index.html": "<header>Home Page</header><main><p>Index Page</p></main><footer>Footer</footer>",
				"about.html": "<header>About Me</header><main><p>About Page</p></main><footer>Footer</footer>",
			},
			expectedErr: nil,
		},
		{
			name: "File copy test",
			templates: map[string]string{
				"pages/robots.txt": "User-agent: *\nDisallow: /",
			},
			data: map[string]interface{}{},
			expectedFiles: map[string]string{
				"robots.txt": "User-agent: *\nDisallow: /",
			},
			expectedErr: nil,
		},
		{
			name: "Pagination test",
			templates: map[string]string{
				"pages/paginated/paginated.tmpl": "<ul>{{range .Iterations.Data}}<li>{{.}}</li>{{end}}</ul>",
			},
			data: map[string]interface{}{
				"pages/paginated/paginated.tmpl.json": map[string]interface{}{
					"Title": "Paginated Data",
					"Iterations": map[string]interface{}{
						"PageSize": 2,
						"PageRoot": "/paginated",
						"Data":     []interface{}{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5"},
					},
				},
			},
			expectedFiles: map[string]string{
				"paginated/paginated.html":   "<ul><li>Item 1</li><li>Item 2</li></ul>",
				"paginated/1/paginated.html": "<ul><li>Item 1</li><li>Item 2</li></ul>",
				"paginated/2/paginated.html": "<ul><li>Item 3</li><li>Item 4</li></ul>",
				"paginated/3/paginated.html": "<ul><li>Item 5</li></ul>",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		tempDir := t.TempDir()
		templateDir := filepath.Join(tempDir, "templates")
		outputDir := filepath.Join(tempDir, "output")

		for path, content := range tc.templates {
			fullPath := filepath.Join(templateDir, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
				t.Errorf("Failed to create test directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte(content), os.ModePerm); err != nil {
				t.Errorf("Failed to create test file: %v", err)
			}
		}

		for path, data := range tc.data {
			fullPath := filepath.Join(templateDir, path)
			jsonData, err := json.Marshal(data)
			if err != nil {
				t.Errorf("Failed to marshal JSON: %v", err)
			}
			if err := os.WriteFile(fullPath, jsonData, os.ModePerm); err != nil {
				t.Errorf("Failed to create JSON data file: %v", err)
			}
		}

		cfg := SiteConfig{
			TemplateDir: templateDir,
			OutputDir:   outputDir,
		}

		gen := NewSiteGenerator(cfg)

		genErr := gen.GenerateSite()
		if (genErr == nil) != (tc.expectedErr == nil) {
			t.Errorf("Test %q failed: expected error %v, got %v", tc.name, tc.expectedErr, genErr)
		}

		if tc.expectedErr != nil {
			if genErr != nil && genErr.Error() != tc.expectedErr.Error() {
				t.Errorf("Test %q failed: expected error %s, got %s", tc.name, tc.expectedErr.Error(), genErr.Error())
			}
		}

		for file, expectedContent := range tc.expectedFiles {
			outputPath := filepath.Join(outputDir, file)
			data, err := os.ReadFile(outputPath)
			if err != nil {
				fileDir := filepath.Dir(outputPath)
				var dirFiles []string
				files, err := os.ReadDir(fileDir)
				if err == nil {
					for _, file := range files {
						dirFiles = append(dirFiles, file.Name())
					}
				}

				t.Errorf("Expected file %s not found. %s: %s", file, fileDir, strings.Join(dirFiles, ", "))
				continue
			}
			if strings.TrimSpace(string(data)) != expectedContent {
				t.Errorf("File %s content mismatch\nExpected:\n%s\nGot:\n%s", file, expectedContent, strings.TrimSpace(string(data)))
			}
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
