package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain entity.
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser is the data required to create a User.
type NewUser struct {
	Email string
	Name  string
}

// UpdateUser is the data allowed when updating a User. Nil fields are left
// unchanged.
type UpdateUser struct {
	Email *string
	Name  *string
}
