package auth

import (
	"fmt"
	"strings"
	"sync"

	"coldchain-alert/internal/domain"
)

type User struct {
	Username string
	Display  string
	Role     domain.Role
}

type Manager struct {
	mu       sync.RWMutex
	users    map[string]User
	sessions map[string]domain.Session
}

func NewManager() *Manager {
	users := map[string]User{
		"warehouse": {Username: "warehouse", Display: "Warehouse Operator", Role: domain.RoleWarehouse},
		"quality":   {Username: "quality", Display: "Quality Inspector", Role: domain.RoleQuality},
		"visitor":   {Username: "visitor", Display: "Read Only Visitor", Role: domain.RoleVisitor},
	}
	return &Manager{users: users, sessions: make(map[string]domain.Session)}
}

func (m *Manager) Login(username string, role domain.Role) (domain.Session, error) {
	username = strings.TrimSpace(username)
	if err := domain.ValidateRole(role); err != nil {
		return domain.Session{}, err
	}
	if username == "" {
		return domain.Session{}, fmt.Errorf("%w: username is required", domain.ErrInvalidInput)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.users[username]
	if !ok || user.Role != role {
		return domain.Session{}, domain.ErrUnauthorized
	}
	token := fmt.Sprintf("session-%s-%s", username, role)
	session := domain.Session{Token: token, Username: username, Role: role}
	m.sessions[token] = session
	return session, nil
}

func (m *Manager) Authenticate(token string) (domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[strings.TrimSpace(token)]
	if !ok {
		return domain.Session{}, domain.ErrUnauthorized
	}
	return session, nil
}

func (m *Manager) Logout(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[token]; !ok {
		return domain.ErrNotFound
	}
	delete(m.sessions, token)
	return nil
}

func (m *Manager) Can(token string, action func(domain.Role) bool) error {
	session, err := m.Authenticate(token)
	if err != nil {
		return err
	}
	if action == nil || !action(session.Role) {
		return domain.ErrUnauthorized
	}
	return nil
}

func (m *Manager) Users() []User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users
}
