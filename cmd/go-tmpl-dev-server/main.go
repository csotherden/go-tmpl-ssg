package main

import (
	"flag"
	"fmt"
	"github.com/csotherden/go-tmpl-ssg/pkg/generator"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	debounceTimer *time.Timer
	clients       = make(map[*websocket.Conn]bool)
	clientsMutex  sync.Mutex
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	rootDir := flag.String("root", "./output", "Root directory to serve files from")
	sourceDir := flag.String("source", "./templates", "Source directory to watch for changes")
	flag.Parse()

	config := generator.SiteConfig{
		TemplateDir:     *sourceDir,
		OutputDir:       *rootDir,
		GenerateSitemap: true,
		BaseURL:         fmt.Sprintf("http://localhost:%s", *port),
		DevServer:       true,
		DevServerAddr:   fmt.Sprintf("localhost:%s", *port),
	}

	recompileTemplates(config)

	absRoot, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path of root directory: %v", err)
	}

	go watchSourceDir(*sourceDir, config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(absRoot, filepath.Clean(r.URL.Path))
		log.Printf("Serving: %s", filePath)
		http.ServeFile(w, r, filePath)
	})

	http.HandleFunc("/ws", handleWebSocket)

	log.Printf("Serving %s on http://localhost:%s", absRoot, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func watchSourceDir(source string, cfg generator.SiteConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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
			debounceRecompile(cfg)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func debounceRecompile(cfg generator.SiteConfig) {
	if debounceTimer != nil {
		debounceTimer.Stop()
	}
	debounceTimer = time.AfterFunc(3*time.Second, func() {
		recompileTemplates(cfg)
	})
}

func recompileTemplates(cfg generator.SiteConfig) {
	log.Println("Recompiling templates...")

	siteGen := generator.NewSiteGenerator(cfg)
	if err := siteGen.GenerateSite(); err != nil {
		log.Printf("Error generating site: %v\n", err)
	} else {
		log.Println("Site generation completed successfully!")
		notifyClients()
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
	}
}

func notifyClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			log.Println("Error sending reload message:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
