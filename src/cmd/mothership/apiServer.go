package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"src/internal/core"
	"src/internal/ml"
	"src/internal/ts"
)

type APIServer struct {
	mothership *core.MotherShip
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	clientsMu  sync.Mutex
	broadcast  chan interface{}
}

func NewAPIServer(ms *core.MotherShip) *APIServer {
	return &APIServer{
		mothership: ms,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}, 100),
	}
}

func (api *APIServer) Start(port string) {
	router := mux.NewRouter()

	// REST Endpoints
	router.HandleFunc("/api/rovers", api.listRovers).Methods("GET")
	router.HandleFunc("/api/rovers/{id}", api.getRover).Methods("GET")
	router.HandleFunc("/api/missions", api.listMissions).Methods("GET")
	router.HandleFunc("/api/missions/{id}", api.getMission).Methods("GET")
	router.HandleFunc("/api/stats", api.getStats).Methods("GET")

	// WebSocket Endpoints
	router.HandleFunc("/ws/telemetry", api.handleWebSocket)

	// CORS middleware
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(router)

	// Broadcaster goroutine
	go api.broadcaster()

	fmt.Printf("🌐 API Server rodando em http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("❌ Erro ao iniciar API Server:", err)
	}
}

// ==================== REST HANDLERS ====================

func (api *APIServer) listRovers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rovers := api.mothership.RoverInfo.ListRovers()
	json.NewEncoder(w).Encode(rovers)
}

func (api *APIServer) getRover(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	rover := api.mothership.RoverInfo.GetRover(uint8(id))
	if rover == nil {
		http.Error(w, "Rover não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(rover)
}

func (api *APIServer) listMissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	missions := api.mothership.MissionManager.ListMissions()
	json.NewEncoder(w).Encode(missions)
}

func (api *APIServer) getMission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	mission := api.mothership.MissionManager.GetMission(uint16(id))
	if mission == nil {
		http.Error(w, "Missão não encontrada", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(mission)
}

func (api *APIServer) getStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rovers := api.mothership.RoverInfo.ListRovers()
	missions := api.mothership.MissionManager.ListMissions()

	stats := map[string]interface{}{
		"total_rovers":         len(rovers),
		"total_missions":       len(missions),
		"active_rovers":        countActiveRovers(rovers),
		"completed_missions":   countCompletedMissions(missions),
		"pending_missions":     countPendingMissions(missions),
		"missions_in_progress": countInProgressMissions(missions),
	}

	json.NewEncoder(w).Encode(stats)
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

	// Envia snapshot inicial
	api.sendSnapshot(conn)

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

func (api *APIServer) sendSnapshot(conn *websocket.Conn) {
	snapshot := map[string]interface{}{
		"event": "snapshot",
		"data": map[string]interface{}{
			"rovers":   api.mothership.RoverInfo.ListRovers(),
			"missions": api.mothership.MissionManager.ListMissions(),
		},
	}
	conn.WriteJSON(snapshot)
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

// Método público para publicar updates em tempo real
func (api *APIServer) PublishUpdate(event string, data interface{}) {
	api.broadcast <- map[string]interface{}{
		"event": event,
		"data":  data,
	}
}

// ==================== HELPERS ====================

func countActiveRovers(rovers []*ts.RoverInfo) int {
	count := 0
	for _, r := range rovers {
		if r.State == "Active" || r.State == "InMission" {
			count++
		}
	}
	return count
}

func countCompletedMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "Completed" {
			count++
		}
	}
	return count
}

func countPendingMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "Pending" {
			count++
		}
	}
	return count
}

func countInProgressMissions(missions []*ml.MissionState) int {
	count := 0
	for _, m := range missions {
		if m.State == "InProgress" {
			count++
		}
	}
	return count
}
