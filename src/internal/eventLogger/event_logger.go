package eventlogger

import (
    "sync"
    "time"
	"src/internal/api"
)

type Event struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`   // INFO, WARN, ERROR
    Source    string    `json:"source"`  // ML, TS, MissionManager, RoverManager, API
    Message   string    `json:"message"`
    Meta      any       `json:"meta,omitempty"` // dados adicionais (seq, roverID, etc)
}

type EventLogger struct {
    mu      sync.Mutex
    history []Event            // últimos N eventos
	api *api.APIServer
}

func NewEventLogger(size int, api *api.APIServer) *EventLogger {
	return &EventLogger{
		history: make([]Event, 0, size), // manter histórico dos últimos N eventos
		api:    api,
	}
}



func (l *EventLogger) Log(level, source, message string, meta any) {
    evt := Event{
        Timestamp: time.Now(),
        Level:     level,
        Source:    source,
        Message:   message,
        Meta:      meta,
    }

    l.mu.Lock()
    if len(l.history) == cap(l.history) {
        l.history = l.history[1:]
    }
    l.history = append(l.history, evt)
    l.mu.Unlock()

    // Se a API estiver conectada, enviar realtime
    if l.api != nil {
        l.api.PublishUpdate("log",evt)
    }
}

func (l *EventLogger) GetHistory() []Event {
    l.mu.Lock()
    defer l.mu.Unlock()

    copyHistory := make([]Event, len(l.history))
    copy(copyHistory, l.history)
    return copyHistory
}

