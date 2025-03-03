package generator

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GenerateSite(templateDir, outputDir string) error {
	componentDir := filepath.Join(templateDir, "components")
	pageDir := filepath.Join(templateDir, "pages")

	componentTemplates, err := template.ParseGlob(filepath.Join(componentDir, "*.tmpl"))
	if err != nil {
		return err
	}

	err = filepath.Walk(pageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(pageDir, path)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(outputDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(outputPath, os.ModePerm)
		}

		if strings.HasSuffix(info.Name(), ".tmpl") {
			return renderTemplate(path, outputPath[:len(outputPath)-5]+".html", componentTemplates)
		} else {
			return copyFile(path, outputPath)
		}
	})

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
