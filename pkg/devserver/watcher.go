package devserver

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/csotherden/go-tmpl-ssg/pkg/generator"
	"github.com/fsnotify/fsnotify"
)

// WatcherConfig holds configuration for the file watcher
type WatcherConfig struct {
	SourceDir   string
	SiteConfig  generator.SiteConfig
	OnRecompile func() error
}

// debounceTimer is used to prevent multiple rapid recompilations
var debounceTimer *time.Timer

// WatchSourceDir watches the source directory for changes and triggers recompilation
func (s *Server) WatchSourceDir() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = filepath.Walk(s.Config.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return watcher.Add(path)
	})
	if err != nil {
		log.Fatalf("Failed to watch source directory: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if strings.HasSuffix(event.Name, "~") {
				continue
			}

			log.Printf("File changed: %s", event.Name)
			s.DebounceRecompile()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// DebounceRecompile debounces multiple recompilation requests
func (s *Server) DebounceRecompile() {
	if debounceTimer != nil {
		debounceTimer.Stop()
	}
	debounceTimer = time.AfterFunc(3*time.Second, func() {
		if err := s.RecompileTemplates(); err != nil {
			log.Printf("Error recompiling templates: %v", err)
		}
	})
}

// RecompileTemplates regenerates the site from templates
func (s *Server) RecompileTemplates() error {
	log.Println("Recompiling templates...")

	siteGen := generator.NewSiteGenerator(s.Config.SiteConfig)
	if err := siteGen.GenerateSite(); err != nil {
		log.Printf("Error generating site: %v\n", err)
		return err
	}

	log.Println("Site generation completed successfully!")
	s.NotifyClients()
	return nil
}
