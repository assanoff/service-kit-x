// Package userdb is the Postgres implementation of user.Store. It maps between
// the domain User and its database row and uses the servicekit sqldb helpers for
// query logging and error translation, following the SDK pg-store convention
// (inline const queries, a model.go row type, dialect-composed pagination).
package userdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/assanoff/servicekit/logger"
	"github.com/assanoff/servicekit/page"
	"github.com/assanoff/servicekit/sqldb"
	"github.com/assanoff/servicekit/sqldb/dialect"

	"github.com/assanoff/service-kit-x/core/user"
)

// Store implements user.Store against Postgres.
type Store struct {
	log     *logger.Logger
	db      *sqlx.DB
	dialect dialect.Dialect
}

// Option customizes a Store.
type Option func(*Store)

// WithDialect overrides the SQL dialect used to compose engine-specific SQL
// (pagination). Defaults to dialect.Postgres.
func WithDialect(d dialect.Dialect) Option {
	return func(s *Store) {
		s.dialect = d
	}
}

// NewStore builds a Store over the connection pool.
func NewStore(log *logger.Logger, db *sqlx.DB, opts ...Option) *Store {
	s := &Store{log: log, db: db, dialect: dialect.Postgres{}}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Compile-time check that Store satisfies the domain contract.
var _ user.Store = (*Store)(nil)

// Create implements user.Store.
func (s *Store) Create(ctx context.Context, u user.User) error {
	const q = `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES (:id, :email, :name, :created_at, :updated_at)`
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(u)); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}

// Update implements user.Store.
func (s *Store) Update(ctx context.Context, u user.User) error {
	const q = `
		UPDATE users
		SET email = :email, name = :name, updated_at = :updated_at
		WHERE id = :id`
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(u)); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// Delete implements user.Store.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = :id`
	data := struct {
		ID string `db:"id"`
	}{ID: id.String()}
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// QueryByID implements user.Store.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	const q = `SELECT id, email, name, created_at, updated_at FROM users WHERE id = :id`
	data := struct {
		ID string `db:"id"`
	}{ID: id.String()}

	var row dbUser
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		return user.User{}, fmt.Errorf("querybyid: %w", err)
	}
	return toCoreUser(row), nil
}

// Count implements user.Store.
func (s *Store) Count(ctx context.Context) (int, error) {
	const q = `SELECT count(*) AS n FROM users`
	var row struct {
		N int `db:"n"`
	}
	if err := sqldb.QueryStruct(ctx, s.log, s.db, q, &row); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return row.N, nil
}

// Query implements user.Store. The pagination clause is composed via the store's
// dialect, which keeps the engine-specific paging syntax behind one seam; it
// binds :offset and :rows_per_page supplied below.
func (s *Store) Query(ctx context.Context, pg page.Page) ([]user.User, error) {
	var buf bytes.Buffer
	buf.WriteString(`SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC`)
	s.dialect.Paginate(&buf)

	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{Offset: pg.Offset(), RowsPerPage: pg.RowsPerPage()}

	var rows []dbUser
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &rows); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return toCoreUsers(rows), nil
}
