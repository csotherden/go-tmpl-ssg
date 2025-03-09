package devserver

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/csotherden/go-tmpl-ssg/pkg/generator"
	"github.com/gorilla/websocket"
)

// Server represents a development server for the static site generator
type Server struct {
	Config       ServerConfig
	clients      map[*websocket.Conn]bool
	clientsMutex sync.Mutex
}

// ServerConfig holds configuration for the development server
type ServerConfig struct {
	Port       string
	RootDir    string
	SourceDir  string
	LiveReload bool
	SiteConfig generator.SiteConfig
}

// NewServer creates a new development server with the given configuration
func NewServer(config ServerConfig) *Server {
	return &Server{
		Config:  config,
		clients: make(map[*websocket.Conn]bool),
	}
}

// Start initializes and starts the development server
func (s *Server) Start() error {
	// Generate the site initially
	if err := s.RecompileTemplates(); err != nil {
		return err
	}

	// Start watching for file changes
	go s.WatchSourceDir()

	// Set up HTTP routes
	absRoot, err := filepath.Abs(s.Config.RootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of root directory: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(absRoot, filepath.Clean(r.URL.Path))
		log.Printf("Serving: %s", filePath)
		http.ServeFile(w, r, filePath)
	})

	// Setup websocket handler for live reload only if enabled
	if s.Config.LiveReload {
		http.HandleFunc("/ws", s.handleWebSocket)
		log.Println("Live reload enabled. Connect to the /ws endpoint for real-time updates.")
	} else {
		log.Println("Live reload disabled. Manual refresh required after file changes.")
	}

	// Start HTTP server
	log.Printf("Serving %s on http://localhost:%s", absRoot, s.Config.Port)
	return http.ListenAndServe(":"+s.Config.Port, nil)
}

// NotifyClients notifies all connected WebSocket clients to reload the page
func (s *Server) NotifyClients() {
	// Skip notification if live reload is disabled
	if !s.Config.LiveReload {
		return
	}

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			log.Println("Error sending reload message:", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleWebSocket handles WebSocket connections for live reload
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.clientsMutex.Lock()
			delete(s.clients, conn)
			s.clientsMutex.Unlock()
			break
		}
	}
}
