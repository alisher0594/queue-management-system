package models

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// InMemoryQueueModel implements queue operations using in-memory storage
type InMemoryQueueModel struct {
	mu       sync.RWMutex
	entries  map[int]*QueueEntry
	nextID   int
	counters map[ServiceType]int // Track queue numbers per service
}

// NewInMemoryQueueModel creates a new in-memory queue model
func NewInMemoryQueueModel() *InMemoryQueueModel {
	return &InMemoryQueueModel{
		entries:  make(map[int]*QueueEntry),
		nextID:   1,
		counters: make(map[ServiceType]int),
	}
}

// Insert adds a new queue entry
func (m *InMemoryQueueModel) Insert(serviceType ServiceType, phoneNumber string) (*QueueEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate queue number
	m.counters[serviceType]++
	queueNumber := fmt.Sprintf("%s%03d", serviceType, m.counters[serviceType])

	entry := &QueueEntry{
		ID:          m.nextID,
		QueueNumber: queueNumber,
		ServiceType: serviceType,
		PhoneNumber: phoneNumber,
		Status:      StatusActive,
		CreatedAt:   time.Now(),
	}

	m.entries[m.nextID] = entry
	m.nextID++

	return entry, nil
}

// Get retrieves a queue entry by ID
func (m *InMemoryQueueModel) Get(id int) (*QueueEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[id]
	if !exists {
		return nil, ErrNoRecord
	}

	return entry, nil
}

// GetByQueueNumber retrieves a queue entry by queue number
func (m *InMemoryQueueModel) GetByQueueNumber(queueNumber string) (*QueueEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.QueueNumber == queueNumber {
			return entry, nil
		}
	}

	return nil, ErrNoRecord
}

// GetByStatus retrieves all queue entries with a specific status
func (m *InMemoryQueueModel) GetByStatus(status QueueStatus, serviceType *ServiceType) ([]*QueueEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var entries []*QueueEntry
	for _, entry := range m.entries {
		if entry.Status == status {
			if serviceType == nil || entry.ServiceType == *serviceType {
				entries = append(entries, entry)
			}
		}
	}

	// Sort by creation time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})

	return entries, nil
}

// GetNextActive retrieves the next active queue entry for a service type
func (m *InMemoryQueueModel) GetNextActive(serviceType ServiceType) (*QueueEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var oldestEntry *QueueEntry
	for _, entry := range m.entries {
		if entry.ServiceType == serviceType && entry.Status == StatusActive {
			if oldestEntry == nil || entry.CreatedAt.Before(oldestEntry.CreatedAt) {
				oldestEntry = entry
			}
		}
	}

	if oldestEntry == nil {
		return nil, ErrNoRecord
	}

	return oldestEntry, nil
}

// UpdateStatus updates the status of a queue entry
func (m *InMemoryQueueModel) UpdateStatus(id int, status QueueStatus, operatorID *int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[id]
	if !exists {
		return ErrNoRecord
	}

	now := time.Now()
	entry.Status = status
	entry.OperatorID = operatorID

	switch status {
	case StatusProcessing:
		entry.CalledAt = &now
	case StatusPostponed:
		entry.PostponedAt = &now
		entry.PostponeCount++
	case StatusServiced:
		entry.ServicedAt = &now
	}

	return nil
}

// GetAll retrieves all queue entries for a specific date and service type
func (m *InMemoryQueueModel) GetAll(date time.Time, serviceType *ServiceType) ([]*QueueEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var entries []*QueueEntry
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	for _, entry := range m.entries {
		if entry.CreatedAt.After(startOfDay) && entry.CreatedAt.Before(endOfDay) {
			if serviceType == nil || entry.ServiceType == *serviceType {
				entries = append(entries, entry)
			}
		}
	}

	// Sort by creation time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})

	return entries, nil
}

// GetStats retrieves queue statistics
func (m *InMemoryQueueModel) GetStats(serviceType ServiceType, date time.Time) (*QueueStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &QueueStats{
		ServiceType: serviceType,
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalWaitTime time.Duration
	var servicedCount int

	for _, entry := range m.entries {
		if entry.ServiceType == serviceType &&
			entry.CreatedAt.After(startOfDay) &&
			entry.CreatedAt.Before(endOfDay) {

			switch entry.Status {
			case StatusActive:
				stats.TotalActive++
			case StatusProcessing:
				stats.TotalProcessing++
			case StatusPostponed:
				stats.TotalPostponed++
			case StatusServiced:
				stats.TotalServiced++
				if entry.ServicedAt != nil {
					totalWaitTime += entry.ServicedAt.Sub(entry.CreatedAt)
					servicedCount++
				}
			}
		}
	}

	if servicedCount > 0 {
		stats.AverageWaitTime = totalWaitTime.Minutes() / float64(servicedCount)
	}

	return stats, nil
}

// Delete removes a queue entry (for cleanup)
func (m *InMemoryQueueModel) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.entries[id]
	if !exists {
		return ErrNoRecord
	}

	delete(m.entries, id)
	return nil
}
