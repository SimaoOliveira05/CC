package eventlogger

import (
    "sync"
    "time"
	"src/internal/api"
)

// Event represents a logged event in the system
type Event struct {
    Timestamp time.Time `json:"timestamp"` // time of the event
    Level     string    `json:"level"`   // INFO, WARN, ERROR
    Source    string    `json:"source"`  // ML, TS, MissionManager, RoverManager, API
    Message   string    `json:"message"` // descriptive message
    Meta      any       `json:"meta,omitempty"` // additional data (seq, roverID, etc)
}

// EventLogger manages logging of events with history and real-time streaming
type EventLogger struct {
    mu      sync.Mutex        
    history []Event            // last N events
    stream  chan Event         // real-time events
	api *api.APIServer         // reference to API server for real-time updates
}

// NewEventLogger creates a new EventLogger with specified history size
func NewEventLogger(size int, api *api.APIServer) *EventLogger {
	return &EventLogger{
		history: make([]Event, 0, size), // maintain history of last N events
		stream:  make(chan Event, 100),  // buffer for real-time events
		api:    api,
	}
}

// Log adds a new event to the logger
func (l *EventLogger) Log(level, source, message string, meta any) {
    evt := Event{
        Timestamp: time.Now(),
        Level:     level,
        Source:    source,
        Message:   message,
        Meta:      meta,
    }

    // Add to history
    l.mu.Lock()
    if len(l.history) == cap(l.history) {
        l.history = l.history[1:]
    }
    l.history = append(l.history, evt)
    l.mu.Unlock()

    // If the API is connected, send real-time update
    if l.api != nil {
        l.api.PublishUpdate("log",evt)
    }
}

// GetHistory returns a copy of the event history
func (l *EventLogger) GetHistory() []Event {
    l.mu.Lock()
    defer l.mu.Unlock()

    copyHistory := make([]Event, len(l.history))
    copy(copyHistory, l.history)
    return copyHistory
}

