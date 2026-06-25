// Package product is the REST transport for the product module. It demonstrates
// the group-scoped authorization pattern of the servicekit web layer: it
// registers no per-route auth, because the whole product group is guarded by the
// server's Install via router.WithApp(authMW...). Every product endpoint —
// including reads — therefore requires authorization when auth is enabled.
// Contrast the user module, which guards only its writes per route. Handlers are
// rest.HandlerFunc values returning an Encoder (a DTO or an *errs.Error).
package product

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/assanoff/servicekit/errs"
	"github.com/assanoff/servicekit/page"
	"github.com/assanoff/servicekit/query"
	"github.com/assanoff/servicekit/web/rest"

	productcore "github.com/assanoff/service-kit-x/core/product"
)

// Handler exposes product endpoints.
type Handler struct {
	core *productcore.Core
}

// New builds a Handler.
func New(core *productcore.Core) *Handler {
	return &Handler{core: core}
}

// Routes registers the product endpoints through the handle seam. It passes no
// per-route middleware: the server's Install hands it the HandleApp of a group
// already wrapped with auth (api.WithApp(authMW...)), so authorization applies to
// the whole group at once — the group-scoped counterpart of user's per-route auth.
func (h *Handler) Routes(handle rest.Handle) {
	handle("GET /products", h.query)
	handle("GET /products/{id}", h.queryByID)
	handle("POST /products", h.create)
	handle("PUT /products/{id}", h.update)
	handle("DELETE /products/{id}", h.delete)
}

// create creates a single product.
func (h *Handler) create(ctx context.Context, r *http.Request) rest.Encoder {
	var req CreateProductReq
	if err := rest.Decode(r, &req); err != nil {
		return errs.From(err)
	}

	p, err := h.core.Create(ctx, productcore.NewProduct{Name: req.Name, Price: req.Price})
	if err != nil {
		return errs.From(err)
	}
	return rest.JSONStatus(toResponse(p), http.StatusCreated)
}

// query lists one page of products (?page, ?rows).
func (h *Handler) query(ctx context.Context, r *http.Request) rest.Encoder {
	pg, err := page.Parse(r.URL.Query().Get("page"), r.URL.Query().Get("rows"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	ps, err := h.core.Query(ctx, pg)
	if err != nil {
		return errs.From(err)
	}
	total, err := h.core.Count(ctx)
	if err != nil {
		return errs.From(err)
	}
	return query.NewResult(toResponseList(ps), total, pg)
}

// queryByID returns one product by id.
func (h *Handler) queryByID(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(rest.Param(r, "id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", rest.Param(r, "id"))
	}
	p, err := h.core.QueryByID(ctx, id)
	if err != nil {
		return errs.From(err)
	}
	return toResponse(p)
}

// update applies a partial update to a product.
func (h *Handler) update(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(rest.Param(r, "id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", rest.Param(r, "id"))
	}

	var req UpdateProductReq
	if err := rest.Decode(r, &req); err != nil {
		return errs.From(err)
	}

	p, err := h.core.Update(ctx, id, productcore.UpdateProduct{Name: req.Name, Price: req.Price})
	if err != nil {
		return errs.From(err)
	}
	return rest.JSON(toResponse(p))
}

// delete removes a product.
func (h *Handler) delete(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(rest.Param(r, "id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", rest.Param(r, "id"))
	}
	if err := h.core.Delete(ctx, id); err != nil {
		return errs.From(err)
	}
	return nil // 204 No Content
}
