// Package widget is an example business module. It demonstrates the servicekit
// conventions: the Core holds business logic and depends only on a Store
// interface declared here, while the concrete SQL implementation lives in a
// nested package (widgetdb). This keeps the domain testable and storage-agnostic.
package widget

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/assanoff/servicekit/errs"
	"github.com/assanoff/servicekit/logger"
	"github.com/assanoff/servicekit/outbox"
	"github.com/assanoff/servicekit/sqldb"
)

// EventWidgetCreated is the CloudEvents type published when a widget is created.
const EventWidgetCreated = "widget.created"

// Widget is the domain entity.
type Widget struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewWidget is the data required to create a Widget.
type NewWidget struct {
	Name        string
	Description string
}

// UpdateWidget is the data allowed when updating a Widget. Nil fields are left
// unchanged.
type UpdateWidget struct {
	Name        *string
	Description *string
}

// Store is the persistence contract for widgets. The Core depends on this
// interface; concrete implementations (e.g. widgetdb) live elsewhere. WithTx
// yields a sibling bound to a transaction so a write can commit atomically with
// an outbox event (see Create).
type Store interface {
	WithTx(tx sqlx.ExtContext) Store
	Create(ctx context.Context, w Widget) error
	Update(ctx context.Context, w Widget) error
	Delete(ctx context.Context, id uuid.UUID) error
	QueryByID(ctx context.Context, id uuid.UUID) (Widget, error)
	Query(ctx context.Context) ([]Widget, error)
}

// Core implements the widget business logic.
type Core struct {
	log   *logger.Logger
	store Store

	// Optional transactional-outbox wiring. When set, Create persists the widget
	// and a widget.created event in one transaction. When nil, Create just writes
	// the widget (no event) — so the example runs without a broker too.
	db     *sqlx.DB
	outbox outbox.Store
	reg    *outbox.Registry
}

// Option customizes a Core.
type Option func(*Core)

// WithOutbox enables transactional event publishing: Create writes the widget
// and a widget.created event atomically via outbox.WithinTran. The event's
// transport route (topic/key) is resolved from reg, registered at startup — the
// domain code names neither an exchange nor a routing key.
func WithOutbox(db *sqlx.DB, store outbox.Store, reg *outbox.Registry) Option {
	return func(c *Core) {
		c.db = db
		c.outbox = store
		c.reg = reg
	}
}

// NewCore constructs a Core.
func NewCore(log *logger.Logger, store Store, opts ...Option) *Core {
	c := &Core{log: log, store: store}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Create validates and persists a new widget. With outbox wiring it also emits
// a widget.created event in the same transaction (transactional outbox); without
// it, the widget is written directly.
func (c *Core) Create(ctx context.Context, nw NewWidget) (Widget, error) {
	now := time.Now().UTC()
	w := Widget{
		ID:          uuid.New(),
		Name:        nw.Name,
		Description: nw.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	var err error
	if c.outbox == nil {
		err = c.store.Create(ctx, w)
	} else {
		// The domain emits a plain typed event; it knows nothing about the
		// transaction's mechanics or the transport. pub records it in the same
		// transaction as the widget write, so both commit (or roll back) together.
		err = outbox.WithinTran(ctx, c.log, c.db, c.outbox, c.reg, func(tx *sqlx.Tx, pub outbox.Publisher) error {
			if cerr := c.store.WithTx(tx).Create(ctx, w); cerr != nil {
				return cerr
			}
			return pub.Publish(ctx, Created{
				ID:          w.ID.String(),
				Name:        w.Name,
				Description: w.Description,
				CreatedAt:   w.CreatedAt,
			})
		})
	}
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return Widget{}, errs.Newf(errs.AlreadyExists, "widget %q already exists", nw.Name).
				WithMessageID("widget.already_exists").
				WithArgs(map[string]any{"name": nw.Name})
		}
		return Widget{}, errs.New(errs.Internal, err)
	}
	return w, nil
}

// Created is the domain event emitted when a widget is created. It is a
// plain payload type: the domain publishes it through outbox.Publisher and the
// Registry (wired at startup) maps it to its transport route. Register it once
// with outbox.Register[Created](reg, EventWidgetCreated, topic, ...).
type Created struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// QueryByID returns a widget or a NotFound error.
func (c *Core) QueryByID(ctx context.Context, id uuid.UUID) (Widget, error) {
	w, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return Widget{}, errs.Newf(errs.NotFound, "widget %s not found", id).
				WithMessageID("widget.not_found").
				WithArgs(map[string]any{"id": id.String()})
		}
		return Widget{}, errs.New(errs.Internal, err)
	}
	return w, nil
}

// Query returns all widgets.
func (c *Core) Query(ctx context.Context) ([]Widget, error) {
	ws, err := c.store.Query(ctx)
	if err != nil {
		return nil, errs.New(errs.Internal, err)
	}
	return ws, nil
}

// Update applies a partial update and persists it.
func (c *Core) Update(ctx context.Context, id uuid.UUID, uw UpdateWidget) (Widget, error) {
	w, err := c.QueryByID(ctx, id)
	if err != nil {
		return Widget{}, err
	}

	if uw.Name != nil {
		w.Name = *uw.Name
	}
	if uw.Description != nil {
		w.Description = *uw.Description
	}
	w.UpdatedAt = time.Now().UTC()

	if err := c.store.Update(ctx, w); err != nil {
		return Widget{}, errs.New(errs.Internal, err)
	}
	return w, nil
}

// Delete removes a widget.
func (c *Core) Delete(ctx context.Context, id uuid.UUID) error {
	if err := c.store.Delete(ctx, id); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return errs.Newf(errs.NotFound, "widget %s not found", id).
				WithMessageID("widget.not_found").
				WithArgs(map[string]any{"id": id.String()})
		}
		return errs.New(errs.Internal, err)
	}
	return nil
}
