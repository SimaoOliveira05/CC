package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// DataProvider is a function that provides data for an API endpoint
type DataProvider func() interface{}

// APIServer represents the API server with REST and WebSocket capabilities.
type APIServer struct {
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	broadcast chan interface{}

	// Map of endpoints to their data provider functions
	dataProviders map[string]DataProvider
	router        *mux.Router
}

// NewAPIServer creates and initializes a new APIServer instance
func NewAPIServer() *APIServer {
	return &APIServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:       make(map[*websocket.Conn]bool),
		broadcast:     make(chan interface{}, 100),
		dataProviders: make(map[string]DataProvider),
		router:        mux.NewRouter(),
	}
	
}

// RegisterEndpoint registers a REST endpoint with a data provider
func (api *APIServer) RegisterEndpoint(path string, method string, provider DataProvider) {
	api.dataProviders[path] = provider
	api.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := provider()
		json.NewEncoder(w).Encode(data)
	}).Methods(method)
}

// RegisterEndpointWithParams registers a REST endpoint with URL parameters
func (api *APIServer) RegisterEndpointWithParams(path string, method string, handler http.HandlerFunc) {
	api.router.HandleFunc(path, handler).Methods(method)
}

// Start starts the API server on the specified port
func (api *APIServer) Start(port string) {
	// WebSocket Endpoint
	api.router.HandleFunc("/ws/telemetry", api.handleWebSocket)

	// CORS middleware
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(api.router)

	// Broadcaster goroutine
	go api.broadcaster()

	fmt.Printf("🌐 API Server running at http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("❌ Error starting API Server:", err)
	}
}

// ==================== WEBSOCKET ====================

// handleWebSocket handles WebSocket connections for real-time updates
func (api *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("❌ Error upgrading WebSocket connection:", err)
		return
	}

	api.clientsMu.Lock()
	api.clients[conn] = true
	api.clientsMu.Unlock()

	fmt.Println("🟢 WebSocket client connected")

	// Keeps the connection open and reads messages (ping/pong)
	defer func() {
		api.clientsMu.Lock()
		delete(api.clients, conn)
		api.clientsMu.Unlock()
		conn.Close()
		fmt.Println("🔴 WebSocket client disconnected")
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// broadcaster sends messages to all connected WebSocket clients
func (api *APIServer) broadcaster() {
	for {
		msg := <-api.broadcast
		api.clientsMu.Lock()
		for client := range api.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(api.clients, client)
			}
		}
		api.clientsMu.Unlock()
	}
}

// PublishUpdate sends a real-time update to all connected WebSocket clients
func (api *APIServer) PublishUpdate(event string, data interface{}) {
	api.broadcast <- map[string]interface{}{
		"event": event,
		"data":  data,
	}
}
