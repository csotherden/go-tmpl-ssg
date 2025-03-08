package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/csotherden/go-tmpl-ssg/pkg/devserver"
	"github.com/csotherden/go-tmpl-ssg/pkg/generator"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	rootDir := flag.String("root", "./output", "Root directory to serve files from")
	sourceDir := flag.String("source", "./templates", "Source directory to watch for changes")
	flag.Parse()

	// Create site configuration
	siteConfig := generator.SiteConfig{
		TemplateDir:     *sourceDir,
		OutputDir:       *rootDir,
		GenerateSitemap: true,
		BaseURL:         fmt.Sprintf("http://localhost:%s", *port),
		DevServer:       true,
		DevServerAddr:   fmt.Sprintf("localhost:%s", *port),
	}

	// Create server configuration
	serverConfig := devserver.ServerConfig{
		Port:       *port,
		RootDir:    *rootDir,
		SourceDir:  *sourceDir,
		SiteConfig: siteConfig,
	}

	// Create and start the development server
	server := devserver.NewServer(serverConfig)
	log.Fatal(server.Start())
}
