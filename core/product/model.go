package product

import (
	"time"

	"github.com/google/uuid"
)

// Product is the domain entity. Price is stored as an integer minor unit (e.g.
// cents) to avoid floating-point money.
type Product struct {
	ID        uuid.UUID
	Name      string
	Price     int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProduct is the data required to create a Product.
type NewProduct struct {
	Name  string
	Price int64
}

// UpdateProduct is the data allowed when updating a Product. Nil fields are left
// unchanged.
type UpdateProduct struct {
	Name  *string
	Price *int64
}
