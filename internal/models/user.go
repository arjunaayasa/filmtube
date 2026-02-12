package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user permissions
type UserRole string

const (
	RoleUser    UserRole = "USER"
	RoleCreator UserRole = "CREATOR"
	RoleAdmin  UserRole = "ADMIN"
)

// User represents a platform user
type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
	Role      UserRole  `db:"role" json:"role"`
	Name      string    `db:"name" json:"name"`
	AvatarURL string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Bio       string    `db:"bio" json:"bio,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
