package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/yosssi/gohtml"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var parentTemplateRegex = regexp.MustCompile(`(?m)^{{-? /\* layout:(.*?) \*/ -?}}`)

type SiteConfig struct {
	TemplateDir     string
	OutputDir       string
	GenerateSitemap bool
	BaseURL         string
	DevServer       bool
	DevServerAddr   string
}

type SiteGenerator struct {
	Config  SiteConfig
	Layouts map[string]string
}

func NewSiteGenerator(config SiteConfig) *SiteGenerator {
	return &SiteGenerator{
		Config:  config,
		Layouts: make(map[string]string),
	}
}

func (s *SiteGenerator) GenerateSite() error {
	pageDir := filepath.Join(s.Config.TemplateDir, "pages")

	componentTemplates, err := s.loadComponents()
	if err != nil {
		return err
	}

	if err := s.loadLayouts(); err != nil {
		return err
	}

	var htmlFiles []string

	err = filepath.Walk(pageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(pageDir, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(s.Config.OutputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(outputPath, os.ModePerm)
		}

		if strings.HasSuffix(info.Name(), ".tmpl") {
			htmlPath := outputPath[:len(outputPath)-5] + ".html"
			htmlFiles = append(htmlFiles, htmlPath)

			jsonPath := path + ".json"
			data := make(map[string]interface{})

			if jsonContent, err := os.ReadFile(jsonPath); err == nil {
				if err := json.Unmarshal(jsonContent, &data); err != nil {
					return err
				}
			}

			if iterations, ok := data["Iterations"].(map[string]interface{}); ok {
				return s.renderPaginatedPages(path, componentTemplates, data, iterations)
			}

			return s.renderPageTemplate(path, htmlPath, componentTemplates, data)
		} else if !strings.HasSuffix(info.Name(), ".tmpl.json") {
			if strings.HasSuffix(info.Name(), ".html") {
				htmlFiles = append(htmlFiles, outputPath)
			}
			return copyFile(path, outputPath)
		}

		return nil
	})

	if err == nil && s.Config.GenerateSitemap {
		err = generateSitemap(s.Config.OutputDir, s.Config.BaseURL, htmlFiles)
	}

	return err
}

func (s *SiteGenerator) loadComponents() (*template.Template, error) {
	componentDir := filepath.Join(s.Config.TemplateDir, "components")

	componentTemplates := template.New("components").Funcs(FuncMap())

	if exists, _ := dirExists(componentDir); !exists {
		return componentTemplates, nil
	}

	// Walk the components directory to load all template files recursively
	err := filepath.Walk(componentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tmpl") {
			relPath, err := filepath.Rel(componentDir, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath) // Ensure Unix-style paths
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(content) == 0 {
				return fmt.Errorf("template %s is empty", relPath)
			}
			_, err = componentTemplates.New(relPath).Parse(string(content))
			if err != nil {
				return fmt.Errorf("error parsing template %s: %v", relPath, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return componentTemplates, nil
}

func (s *SiteGenerator) loadLayouts() error {
	layoutDir := filepath.Join(s.Config.TemplateDir, "layouts")

	if exists, _ := dirExists(layoutDir); !exists {
		return nil
	}

	files, err := os.ReadDir(layoutDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tmpl") {
			path := filepath.Join(layoutDir, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			s.Layouts[file.Name()] = string(content)
		}
	}
	return nil
}

func (s *SiteGenerator) renderPaginatedPages(pagePath string, componentTemplates *template.Template, data map[string]interface{}, iterations map[string]interface{}) error {
	pageSize, ok := iterations["PageSize"].(float64)
	if !ok || pageSize <= 0 {
		return fmt.Errorf("invalid PageSize in Iterations for %s", pagePath)
	}

	pageRoot, _ := iterations["PageRoot"].(string)

	items, ok := iterations["Data"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid Data array in Iterations for %s", pagePath)
	}

	totalItems := len(items)
	totalPages := (totalItems + int(pageSize) - 1) / int(pageSize)

	outputBaseDir := s.Config.OutputDir

	for pageNumber := 1; pageNumber <= totalPages; pageNumber++ {
		start := (pageNumber - 1) * int(pageSize)
		end := start + int(pageSize)
		if end > totalItems {
			end = totalItems
		}

		pagedData := make(map[string]interface{})
		for k, v := range data {
			pagedData[k] = v
		}
		pagedData["Iterations"] = map[string]interface{}{
			"PageSize":   pageSize,
			"PageNumber": pageNumber,
			"TotalPages": totalPages,
			"TotalCount": totalItems,
			"PageRoot":   pageRoot,
			"Data":       items[start:end],
		}

		var pageOutputPath string
		relPath, err := filepath.Rel(filepath.Join(s.Config.TemplateDir, "pages"), pagePath)
		if err != nil {
			return fmt.Errorf("error resolving relative path: %v", err)
		}
		relPath = filepath.ToSlash(relPath) // Ensure correct path separators

		pageDir := filepath.Join(outputBaseDir, filepath.Dir(relPath))
		pageName := filepath.Base(pagePath[:len(pagePath)-5])

		err = os.MkdirAll(filepath.Join(pageDir, strconv.Itoa(pageNumber)), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating paginated output directory: %v", err)
		}

		pageOutputPath = filepath.Join(pageDir, strconv.Itoa(pageNumber), pageName+".html")

		err = s.renderPageTemplate(pagePath, pageOutputPath, componentTemplates, pagedData)
		if err != nil {
			return fmt.Errorf("error rendering paginated page %s: %v", pageOutputPath, err)
		}

		if pageNumber == 1 {
			basePageOutputPath := filepath.Join(pageDir, pageName+".html")
			if err := s.renderPageTemplate(pagePath, basePageOutputPath, componentTemplates, pagedData); err != nil {
				return fmt.Errorf("error rendering paginated page %s: %v", basePageOutputPath, err)
			}
		}
	}

	return nil
}

func (s *SiteGenerator) renderPageTemplate(pagePath, outputPath string, componentTemplates *template.Template, data map[string]interface{}) error {
	content, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}

	matches := parentTemplateRegex.FindSubmatch(content)
	if matches != nil {
		parentName := string(matches[1])
		parentContent, exists := s.Layouts[parentName]
		if exists {
			tmpl, err := componentTemplates.Clone()
			if err != nil {
				return err
			}
			childTemplateName := filepath.ToSlash(filepath.Base(pagePath))

			_, err = tmpl.New(childTemplateName).Parse(string(content))
			if err != nil {
				return err
			}

			parentContent = strings.ReplaceAll(parentContent, "{{template %CONTENT% .}}", "{{template \""+childTemplateName+"\" .}}")
			parentContent = strings.ReplaceAll(parentContent, "{{template %CONTENT%}}", "{{template \""+childTemplateName+"\" .}}")
			_, err = tmpl.New(parentName).Parse(parentContent)
			if err != nil {
				return err
			}

			return s.executeTemplate(outputPath, tmpl, parentName, data)
		}
	}

	finalTemplate, err := componentTemplates.Clone()
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(s.Config.TemplateDir, pagePath)
	if err != nil {
		return err
	}
	relPath = filepath.ToSlash(relPath)

	_, err = finalTemplate.New(relPath).Parse(string(content))
	if err != nil {
		return err
	}

	return s.executeTemplate(outputPath, finalTemplate, relPath, data)
}

func (s *SiteGenerator) executeTemplate(outputPath string, tmpl *template.Template, templateName string, data interface{}) error {
	var buf bytes.Buffer

	// Execute the template into the buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return err
	}

	// Convert buffer to string for manipulation
	htmlContent := buf.String()

	// Inject WebSocket script if DevServer is enabled
	if s.Config.DevServer {
		websocketScript := fmt.Sprintf(`<script>
			function connectWebSocket() {
				const socket = new WebSocket("ws://%s/ws");

				socket.onopen = function() {
					console.log("Connected to live reload server.");
				};

				socket.onmessage = function(event) {
					if (event.data === "reload") {
						console.log("Reload event received. Refreshing page...");
						location.reload();
					}
				};

				socket.onerror = function(error) {
					console.error("WebSocket error:", error);
				};

				socket.onclose = function() {
					console.warn("WebSocket connection lost. Reconnecting in 15 seconds...");
					setTimeout(connectWebSocket, 15000);
				};
			}

			connectWebSocket();
		</script>`, s.Config.DevServerAddr)

		// Insert before </head> tag
		htmlContent = strings.Replace(htmlContent, "</head>", websocketScript+"\n</head>", 1)
	}

	// Format the HTML output
	formattedHTML := gohtml.Format(htmlContent)

	// Write the final output to the file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(formattedHTML)
	return err
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return true, nil
		}
		return false, nil
	}
	if os.IsNotExist(err) {
		// Directory does not exist
		return false, nil
	}
	// Error occurred (e.g., permission denied)
	return false, err
}
