package ws

import "sync"

// Room is a single meeting's set of connected clients with fan-out helpers.
type Room struct {
	slug string
	hub  *Hub

	mu      sync.RWMutex
	clients map[string]*Client
}

func newRoom(slug string, hub *Hub) *Room {
	return &Room{
		slug:    slug,
		hub:     hub,
		clients: make(map[string]*Client),
	}
}

// add registers a client and returns the snapshot of pre-existing peers.
func (r *Room) add(c *Client) []PeerInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	peers := make([]PeerInfo, 0, len(r.clients))
	for _, existing := range r.clients {
		peers = append(peers, PeerInfo{ID: existing.id, Name: existing.name})
	}
	r.clients[c.id] = c
	return peers
}

// remove deletes a client; the hub cleans up empty rooms.
func (r *Room) remove(id string) {
	r.mu.Lock()
	delete(r.clients, id)
	r.mu.Unlock()
	r.hub.removeRoom(r.slug)
}

func (r *Room) isEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients) == 0
}

// broadcast sends data to all clients except exceptID ("" sends to everyone).
func (r *Room) broadcast(data []byte, exceptID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, c := range r.clients {
		if id == exceptID {
			continue
		}
		c.trySend(data)
	}
}

// sendTo delivers data to a single client by id; returns false if not present.
func (r *Room) sendTo(id string, data []byte) bool {
	r.mu.RLock()
	c, ok := r.clients[id]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	c.trySend(data)
	return true
}
