package sse

import (
	"sync"
)

type Conn struct {
	ID     string
	UserID string
	ListID string
	Events chan Event
	done   chan struct{}

	mu             sync.Mutex
	correlationIDs map[string]struct{}
}

type Event struct {
	Name string // SSE event name: "item.created", "item.toggled", "item.deleted"
	Data string // HTML fragment
}

type Manager struct {
	mu          sync.Mutex
	connections map[string]*Conn   // conn ID → conn
	byUser      map[string][]*Conn // user ID → conns
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Conn),
		byUser:      make(map[string][]*Conn),
	}
}

func (m *Manager) Register(conn *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[conn.ID] = conn
	m.byUser[conn.UserID] = append(m.byUser[conn.UserID], conn)
}

func (m *Manager) Unregister(conn *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, conn.ID)
	conns := m.byUser[conn.UserID]
	for i, c := range conns {
		if c.ID == conn.ID {
			m.byUser[conn.UserID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(m.byUser[conn.UserID]) == 0 {
		delete(m.byUser, conn.UserID)
	}
	close(conn.done)
}

// AddCorrelationID registers a correlation ID with all SSE connections for a user.
func (m *Manager) AddCorrelationID(userID, correlationID string) {
	m.mu.Lock()
	conns := m.byUser[userID]
	m.mu.Unlock()

	for _, conn := range conns {
		conn.mu.Lock()
		conn.correlationIDs[correlationID] = struct{}{}
		conn.mu.Unlock()
	}
}

// HasCorrelationID checks if a connection originated the given correlation ID.
// If found, it removes the ID (one-time use).
func (conn *Conn) HasCorrelationID(id string) bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if _, ok := conn.correlationIDs[id]; ok {
		delete(conn.correlationIDs, id)
		return true
	}
	return false
}

func NewConn(id, userID, listID string) *Conn {
	return &Conn{
		ID:             id,
		UserID:         userID,
		ListID:         listID,
		Events:         make(chan Event, 16),
		done:           make(chan struct{}),
		correlationIDs: make(map[string]struct{}),
	}
}

func (conn *Conn) Done() <-chan struct{} {
	return conn.done
}
