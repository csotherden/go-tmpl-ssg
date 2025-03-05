package generator

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var parentTemplateRegex = regexp.MustCompile(`(?m)^{{-? /\* layout:(.*?) \*/ -?}}`)

type SiteConfig struct {
	TemplateDir     string
	OutputDir       string
	GenerateSitemap bool
	BaseURL         string
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
	componentDir := filepath.Join(s.Config.TemplateDir, "components")
	layoutDir := filepath.Join(s.Config.TemplateDir, "layouts")
	pageDir := filepath.Join(s.Config.TemplateDir, "pages")

	componentTemplates := template.New("components").Funcs(FuncMap())
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
		return err
	}

	if err := s.loadLayouts(layoutDir); err != nil {
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
			return s.renderPageTemplate(path, htmlPath, componentTemplates)
		} else {
			if strings.HasSuffix(info.Name(), ".html") {
				htmlFiles = append(htmlFiles, outputPath)
			}
			return copyFile(path, outputPath)
		}
	})

	if err == nil && s.Config.GenerateSitemap {
		err = generateSitemap(s.Config.OutputDir, s.Config.BaseURL, htmlFiles)
	}

	return err
}

func (s *SiteGenerator) loadLayouts(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tmpl") {
			path := filepath.Join(dir, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			s.Layouts[file.Name()] = string(content)
		}
	}
	return nil
}

func (s *SiteGenerator) renderPageTemplate(pagePath, outputPath string, componentTemplates *template.Template) error {
	content, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	jsonPath := pagePath + ".json"
	if jsonContent, err := os.ReadFile(jsonPath); err == nil {
		if err := json.Unmarshal(jsonContent, &data); err != nil {
			return err
		}
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

			return executeTemplate(outputPath, tmpl, parentName, data)
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

	return executeTemplate(outputPath, finalTemplate, relPath, data)
}

func executeTemplate(outputPath string, tmpl *template.Template, templateName string, data interface{}) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return tmpl.ExecuteTemplate(outputFile, templateName, data)
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
