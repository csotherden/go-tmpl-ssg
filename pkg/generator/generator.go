package generator

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type SiteConfig struct {
	TemplateDir     string
	OutputDir       string
	GenerateSitemap bool
	BaseURL         string
}

type SiteGenerator struct {
	Config SiteConfig
}

func NewSiteGenerator(config SiteConfig) *SiteGenerator {
	return &SiteGenerator{Config: config}
}

func (s *SiteGenerator) GenerateSite() error {
	componentDir := filepath.Join(s.Config.TemplateDir, "components")
	pageDir := filepath.Join(s.Config.TemplateDir, "pages")

	componentTemplates, err := template.ParseGlob(filepath.Join(componentDir, "*.tmpl"))
	if err != nil {
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
			return renderTemplate(path, htmlPath, componentTemplates)
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

func renderTemplate(templatePath, outputPath string, componentTemplates *template.Template) error {
	pageTemplate, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	finalTemplate, err := componentTemplates.Clone()
	if err != nil {
		return err
	}

	finalTemplate, err = finalTemplate.AddParseTree("page", pageTemplate.Tree)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return finalTemplate.ExecuteTemplate(outputFile, "page", nil)
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
