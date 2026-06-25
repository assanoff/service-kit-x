// Package product is an example business module: a clean REST CRUD vertical that
// demonstrates the servicekit web architecture. The Core holds business logic
// and depends only on the Store interface declared here; the Postgres
// implementation lives in the nested productdb package.
package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/assanoff/servicekit/errs"
	"github.com/assanoff/servicekit/logger"
	"github.com/assanoff/servicekit/page"
	"github.com/assanoff/servicekit/sqldb"
)

// Store is the persistence contract for products.
type Store interface {
	Create(ctx context.Context, p Product) error
	Update(ctx context.Context, p Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	QueryByID(ctx context.Context, id uuid.UUID) (Product, error)
	Query(ctx context.Context, pg page.Page) ([]Product, error)
	Count(ctx context.Context) (int, error)
}

// Core implements the product business logic.
type Core struct {
	log   *logger.Logger
	store Store
}

// NewCore constructs a Core.
func NewCore(log *logger.Logger, store Store) *Core {
	return &Core{log: log, store: store}
}

// Create validates and persists a new product.
func (c *Core) Create(ctx context.Context, np NewProduct) (Product, error) {
	now := time.Now().UTC()
	p := Product{
		ID:        uuid.New(),
		Name:      np.Name,
		Price:     np.Price,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := c.store.Create(ctx, p); err != nil {
		return Product{}, errs.New(errs.Internal, err)
	}
	return p, nil
}

// QueryByID returns a product or a NotFound error.
func (c *Core) QueryByID(ctx context.Context, id uuid.UUID) (Product, error) {
	p, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Product{}, c.notFound(id)
		}
		return Product{}, errs.New(errs.Internal, err)
	}
	return p, nil
}

// Query returns one page of products, newest first.
func (c *Core) Query(ctx context.Context, pg page.Page) ([]Product, error) {
	ps, err := c.store.Query(ctx, pg)
	if err != nil {
		return nil, errs.New(errs.Internal, err)
	}
	return ps, nil
}

// Count returns the total number of products.
func (c *Core) Count(ctx context.Context) (int, error) {
	n, err := c.store.Count(ctx)
	if err != nil {
		return 0, errs.New(errs.Internal, err)
	}
	return n, nil
}

// Update applies a partial update and persists it.
func (c *Core) Update(ctx context.Context, id uuid.UUID, up UpdateProduct) (Product, error) {
	p, err := c.QueryByID(ctx, id)
	if err != nil {
		return Product{}, err
	}
	if up.Name != nil {
		p.Name = *up.Name
	}
	if up.Price != nil {
		p.Price = *up.Price
	}
	p.UpdatedAt = time.Now().UTC()

	if err := c.store.Update(ctx, p); err != nil {
		return Product{}, errs.New(errs.Internal, err)
	}
	return p, nil
}

// Delete removes a product.
func (c *Core) Delete(ctx context.Context, id uuid.UUID) error {
	if err := c.store.Delete(ctx, id); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return c.notFound(id)
		}
		return errs.New(errs.Internal, err)
	}
	return nil
}

// notFound builds the canonical not-found error for id.
func (c *Core) notFound(id uuid.UUID) error {
	return errs.Newf(errs.NotFound, "product %s not found", id).
		WithMessageID("product.not_found").
		WithArgs(map[string]any{"id": id.String()})
}
