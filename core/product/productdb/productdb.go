// Package productdb is the Postgres implementation of product.Store, following
// the SDK pg-store convention (inline const queries, a model.go row type,
// dialect-composed pagination) and the servicekit sqldb helpers.
package productdb

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

	"github.com/assanoff/service-kit-x/core/product"
)

// Store implements product.Store against Postgres.
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
var _ product.Store = (*Store)(nil)

// Create implements product.Store.
func (s *Store) Create(ctx context.Context, p product.Product) error {
	const q = `
		INSERT INTO products (id, name, price, created_at, updated_at)
		VALUES (:id, :name, :price, :created_at, :updated_at)`
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(p)); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}

// Update implements product.Store.
func (s *Store) Update(ctx context.Context, p product.Product) error {
	const q = `
		UPDATE products
		SET name = :name, price = :price, updated_at = :updated_at
		WHERE id = :id`
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(p)); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// Delete implements product.Store.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM products WHERE id = :id`
	data := struct {
		ID string `db:"id"`
	}{ID: id.String()}
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// QueryByID implements product.Store.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	const q = `SELECT id, name, price, created_at, updated_at FROM products WHERE id = :id`
	data := struct {
		ID string `db:"id"`
	}{ID: id.String()}

	var row dbProduct
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		return product.Product{}, fmt.Errorf("querybyid: %w", err)
	}
	return toCoreProduct(row), nil
}

// Count implements product.Store.
func (s *Store) Count(ctx context.Context) (int, error) {
	const q = `SELECT count(*) AS n FROM products`
	var row struct {
		N int `db:"n"`
	}
	if err := sqldb.QueryStruct(ctx, s.log, s.db, q, &row); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return row.N, nil
}

// Query implements product.Store. The pagination clause is composed via the
// store's dialect, binding :offset and :rows_per_page supplied below.
func (s *Store) Query(ctx context.Context, pg page.Page) ([]product.Product, error) {
	var buf bytes.Buffer
	buf.WriteString(`SELECT id, name, price, created_at, updated_at FROM products ORDER BY created_at DESC`)
	s.dialect.Paginate(&buf)

	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{Offset: pg.Offset(), RowsPerPage: pg.RowsPerPage()}

	var rows []dbProduct
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &rows); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return toCoreProducts(rows), nil
}
