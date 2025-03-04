package generator

import (
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

	componentTemplates := template.New("components")
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

	matches := parentTemplateRegex.FindSubmatch(content)
	if matches != nil {
		parentName := string(matches[1])
		parentContent, exists := s.Layouts[parentName]
		if exists {
			tmpl, err := componentTemplates.Clone()
			if err != nil {
				return err
			}
			childTemplateName := filepath.Base(pagePath)

			// Define the child template in the template set
			tmpl, err = tmpl.New(childTemplateName).Parse(string(content))
			if err != nil {
				return err
			}

			// Replace %CONTENT% placeholder with actual child template call
			parentContent = strings.Replace(parentContent, "{{template %CONTENT%}}", "{{template \""+childTemplateName+"\" .}}", 1)
			tmpl, err = tmpl.New(parentName).Parse(parentContent)
			if err != nil {
				return err
			}

			return executeTemplate(outputPath, tmpl, parentName)
		}
	}

	finalTemplate, err := componentTemplates.Clone()
	if err != nil {
		return err
	}

	finalTemplate, err = finalTemplate.New(filepath.Base(pagePath)).Parse(string(content))
	if err != nil {
		return err
	}

	return executeTemplate(outputPath, finalTemplate, filepath.Base(pagePath))
}

func executeTemplate(outputPath string, tmpl *template.Template, templateName string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return tmpl.ExecuteTemplate(outputFile, templateName, nil)
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
