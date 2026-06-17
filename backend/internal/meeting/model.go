// Package meeting is the domain for meeting rooms: models, service, and the
// repository interface. It does not import transport or concrete infrastructure.
package meeting

import (
	"errors"
	"time"
)

// Status values for a meeting.
const (
	StatusActive = "active"
	StatusEnded  = "ended"
)

// Domain errors, mapped to HTTP statuses at the transport layer.
var (
	ErrNotFound = errors.New("meeting not found")
)

// Meeting is the core aggregate.
type Meeting struct {
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	Title     string     `json:"title"`
	HostName  string     `json:"host_name"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Participant represents someone who joined a meeting (persisted record).
type Participant struct {
	ID          string     `json:"id"`
	MeetingID   string     `json:"meeting_id"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	JoinedAt    time.Time  `json:"joined_at"`
	LeftAt      *time.Time `json:"left_at,omitempty"`
}
