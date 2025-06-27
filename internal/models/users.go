package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// InMemoryUserModel implements user operations using in-memory storage
type InMemoryUserModel struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewInMemoryUserModel creates a new in-memory user model
func NewInMemoryUserModel() *InMemoryUserModel {
	model := &InMemoryUserModel{
		users:  make(map[int]*User),
		nextID: 1,
	}

	// Create default admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	model.users[1] = &User{
		ID:             1,
		Name:           "Administrator",
		Email:          "admin@queue.com",
		HashedPassword: adminPassword,
		Role:           "admin",
		ServiceType:    ServiceTypeA,
		Created:        time.Now(),
		Active:         true,
	}

	// Create default operators
	operatorAPassword, _ := bcrypt.GenerateFromPassword([]byte("operator123"), 12)
	model.users[2] = &User{
		ID:             2,
		Name:           "Operator A",
		Email:          "operatora@queue.com",
		HashedPassword: operatorAPassword,
		Role:           "operator",
		ServiceType:    ServiceTypeA,
		Created:        time.Now(),
		Active:         true,
	}

	operatorBPassword, _ := bcrypt.GenerateFromPassword([]byte("operator123"), 12)
	model.users[3] = &User{
		ID:             3,
		Name:           "Operator B",
		Email:          "operatorb@queue.com",
		HashedPassword: operatorBPassword,
		Role:           "operator",
		ServiceType:    ServiceTypeB,
		Created:        time.Now(),
		Active:         true,
	}

	operatorCPassword, _ := bcrypt.GenerateFromPassword([]byte("operator123"), 12)
	model.users[4] = &User{
		ID:             4,
		Name:           "Operator C",
		Email:          "operatorc@queue.com",
		HashedPassword: operatorCPassword,
		Role:           "operator",
		ServiceType:    ServiceTypeC,
		Created:        time.Now(),
		Active:         true,
	}

	model.nextID = 5

	return model
}

// Insert creates a new user
func (m *InMemoryUserModel) Insert(name, email, password, role string, serviceType ServiceType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if email already exists
	for _, user := range m.users {
		if user.Email == email {
			return ErrDuplicateEmail
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	user := &User{
		ID:             m.nextID,
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
		Role:           role,
		ServiceType:    serviceType,
		Created:        time.Now(),
		Active:         true,
	}

	m.users[m.nextID] = user
	m.nextID++

	return nil
}

// Authenticate verifies user credentials
func (m *InMemoryUserModel) Authenticate(email, password string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email && user.Active {
			err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
			if err != nil {
				return 0, ErrInvalidCredentials
			}
			return user.ID, nil
		}
	}

	return 0, ErrInvalidCredentials
}

// Get retrieves a user by ID
func (m *InMemoryUserModel) Get(id int) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists || !user.Active {
		return nil, ErrNoRecord
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (m *InMemoryUserModel) GetByEmail(email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email && user.Active {
			return user, nil
		}
	}

	return nil, ErrNoRecord
}

// InMemorySessionModel implements session operations using in-memory storage
type InMemorySessionModel struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewInMemorySessionModel creates a new in-memory session model
func NewInMemorySessionModel() *InMemorySessionModel {
	return &InMemorySessionModel{
		sessions: make(map[string]*Session),
	}
}

// Insert creates a new session
func (m *InMemorySessionModel) Insert(userID int, expiry time.Time) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a random session token
	token, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	// Hash the token
	hasher := sha256.New()
	hasher.Write([]byte(token))
	hashedToken := base32.StdEncoding.EncodeToString(hasher.Sum(nil))

	session := &Session{
		Token:  hashedToken,
		UserID: userID,
		Expiry: expiry,
	}

	m.sessions[hashedToken] = session

	return hashedToken, nil
}

// Get retrieves a session by token
func (m *InMemorySessionModel) Get(token string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[token]
	if !exists {
		return 0, ErrNoRecord
	}

	if time.Now().After(session.Expiry) {
		delete(m.sessions, token)
		return 0, ErrNoRecord
	}

	return session.UserID, nil
}

// Delete removes a session
func (m *InMemorySessionModel) Delete(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, token)
	return nil
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(bytes)[:length], nil
}
