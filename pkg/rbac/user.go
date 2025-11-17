package rbac

import (
	"fmt"
	"time"
)

// User represents a containr user
type User struct {
	ID        string         `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email,omitempty"`
	Roles     []*Role        `json:"roles"`
	Quota     *ResourceQuota `json:"quota,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewUser creates a new user
func NewUser(username string) *User {
	return &User{
		ID:        generateID(),
		Username:  username,
		Roles:     make([]*Role, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// HasRole checks if user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(resource string, action Action) bool {
	for _, role := range u.Roles {
		if role.HasPermission(resource, action) {
			return true
		}
	}
	return false
}

// generateID generates a unique ID for users
func generateID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}
