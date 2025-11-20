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

// DataProvider é uma função que retorna dados para um endpoint específico
type DataProvider func() interface{}

type APIServer struct {
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	broadcast chan interface{}

	// Mapa de endpoints para suas funções fornecedoras de dados
	dataProviders map[string]DataProvider
	router        *mux.Router
}

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

// RegisterEndpoint registra um endpoint REST com um provider de dados
func (api *APIServer) RegisterEndpoint(path string, method string, provider DataProvider) {
	api.dataProviders[path] = provider
	api.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := provider()
		json.NewEncoder(w).Encode(data)
	}).Methods(method)
}

// RegisterEndpointWithParams registra um endpoint REST com parâmetros na URL
func (api *APIServer) RegisterEndpointWithParams(path string, method string, handler http.HandlerFunc) {
	api.router.HandleFunc(path, handler).Methods(method)
}

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

	fmt.Printf("🌐 API Server rodando em http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("❌ Erro ao iniciar API Server:", err)
	}
}

// ==================== WEBSOCKET ====================

func (api *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("❌ Erro ao upgradar conexão WebSocket:", err)
		return
	}

	api.clientsMu.Lock()
	api.clients[conn] = true
	api.clientsMu.Unlock()

	fmt.Println("🟢 Cliente WebSocket conectado")

	// Mantém a conexão aberta e lê mensagens (ping/pong)
	defer func() {
		api.clientsMu.Lock()
		delete(api.clients, conn)
		api.clientsMu.Unlock()
		conn.Close()
		fmt.Println("🔴 Cliente WebSocket desconectado")
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

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

// PublishUpdate envia um update em tempo real para todos os clientes WebSocket conectados
func (api *APIServer) PublishUpdate(event string, data interface{}) {
	api.broadcast <- map[string]interface{}{
		"event": event,
		"data":  data,
	}
}
