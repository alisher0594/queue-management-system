package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

// QueueStatus represents the status of a queue entry
type QueueStatus string

const (
	StatusActive     QueueStatus = "active"
	StatusProcessing QueueStatus = "processing"
	StatusServiced   QueueStatus = "serviced"
	StatusPostponed  QueueStatus = "postponed"
)

// ServiceType represents different service types
type ServiceType string

const (
	ServiceTypeA ServiceType = "A"
	ServiceTypeB ServiceType = "B"
	ServiceTypeC ServiceType = "C"
)

// QueueEntry represents a single queue entry
type QueueEntry struct {
	ID            int         `json:"id"`
	QueueNumber   string      `json:"queue_number"`
	ServiceType   ServiceType `json:"service_type"`
	PhoneNumber   string      `json:"phone_number"`
	Status        QueueStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	CalledAt      *time.Time  `json:"called_at"`
	PostponedAt   *time.Time  `json:"postponed_at"`
	ServicedAt    *time.Time  `json:"serviced_at"`
	PostponeCount int         `json:"postpone_count"`
	OperatorID    *int        `json:"operator_id"`
}

// User represents a user (operator or admin)
type User struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Email          string      `json:"email"`
	HashedPassword []byte      `json:"-"`
	Role           string      `json:"role"` // "admin" or "operator"
	ServiceType    ServiceType `json:"service_type"`
	Created        time.Time   `json:"created"`
	Active         bool        `json:"active"`
}

// Session represents a user session
type Session struct {
	Token  string
	UserID int
	Expiry time.Time
}

// QueueStats represents queue statistics
type QueueStats struct {
	ServiceType     ServiceType `json:"service_type"`
	TotalActive     int         `json:"total_active"`
	TotalProcessing int         `json:"total_processing"`
	TotalPostponed  int         `json:"total_postponed"`
	TotalServiced   int         `json:"total_serviced"`
	AverageWaitTime float64     `json:"average_wait_time"`
}
