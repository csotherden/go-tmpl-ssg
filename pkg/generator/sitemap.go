package generator

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

func generateSitemap(outputDir, baseURL string, htmlFiles []string) error {
	type URL struct {
		Loc string `xml:"loc"`
	}
	type URLSet struct {
		XMLName xml.Name `xml:"urlset"`
		URLs    []URL    `xml:"url"`
	}

	var urls []URL
	for _, file := range htmlFiles {
		relPath, _ := filepath.Rel(outputDir, file)
		urls = append(urls, URL{Loc: baseURL + "/" + strings.ReplaceAll(relPath, "\\", "/")})
	}

	urlSet := URLSet{URLs: urls}
	sitemapPath := filepath.Join(outputDir, "sitemap.xml")
	file, err := os.Create(sitemapPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	if err != nil {
		return err
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(urlSet)
}
