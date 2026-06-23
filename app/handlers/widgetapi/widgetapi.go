// Package widgetapi is the REST transport layer for the widget module. Handlers
// are servicekit rest.HandlerFunc values: they decode/validate input, call the
// Core, and return an Encoder (a DTO or an *errs.Error).
package widgetapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/assanoff/service-kit-x/core/widget"
	"github.com/assanoff/service-kit-x/core/widgetimport"
	"github.com/assanoff/servicekit/errs"
	"github.com/assanoff/servicekit/web/rest"
	"github.com/assanoff/servicekit/web/router"
)

// Handler exposes widget endpoints.
type Handler struct {
	core     *widget.Core
	importer *widgetimport.Importer
}

// New builds a Handler.
func New(core *widget.Core, importer *widgetimport.Importer) *Handler {
	return &Handler{core: core, importer: importer}
}

// Routes registers the widget endpoints on r. Reads are public; writes are
// registered on a sub-router carrying any authMW (e.g. JWT auth + RBAC) so they
// require authorization when auth is enabled.
func (h *Handler) Routes(r *router.Router, authMW ...router.Middleware) {
	r.HandleApp("GET /widgets", h.query)
	r.HandleApp("GET /widgets/{id}", h.queryByID)

	w := r
	if len(authMW) > 0 {
		w = r.With(authMW...)
	}
	w.HandleApp("POST /widgets", h.create)
	w.HandleApp("POST /widgets/import", h.importBatch)
	w.HandleApp("PUT /widgets/{id}", h.update)
	w.HandleApp("DELETE /widgets/{id}", h.delete)
}

// importBatch enqueues a batch of widgets for asynchronous bulk insertion by the
// background worker. It returns 202 Accepted immediately; the widgets appear
// once the worker drains the queue.
func (h *Handler) importBatch(ctx context.Context, r *http.Request) rest.Encoder {
	var req importWidgetsReq
	if err := rest.Decode(r, &req); err != nil {
		return errs.From(err)
	}

	news := make([]widget.NewWidget, len(req.Widgets))
	for i, w := range req.Widgets {
		news[i] = widget.NewWidget{Name: w.Name, Description: w.Description}
	}

	scheduled, err := h.importer.Schedule(ctx, req.Name, news)
	if err != nil {
		return errs.From(err)
	}
	return rest.JSONStatus(importResponse{Scheduled: scheduled, Count: len(news)}, http.StatusAccepted)
}

func (h *Handler) create(ctx context.Context, r *http.Request) rest.Encoder {
	var req createWidgetReq
	if err := rest.Decode(r, &req); err != nil {
		return errs.From(err)
	}

	w, err := h.core.Create(ctx, widget.NewWidget{Name: req.Name, Description: req.Description})
	if err != nil {
		return errs.From(err)
	}
	return rest.JSONStatus(toResponse(w), http.StatusCreated)
}

func (h *Handler) query(ctx context.Context, _ *http.Request) rest.Encoder {
	ws, err := h.core.Query(ctx)
	if err != nil {
		return errs.From(err)
	}
	return rest.JSON(toResponses(ws))
}

func (h *Handler) queryByID(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", r.PathValue("id"))
	}

	w, err := h.core.QueryByID(ctx, id)
	if err != nil {
		return errs.From(err)
	}
	return rest.JSON(toResponse(w))
}

func (h *Handler) update(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", r.PathValue("id"))
	}

	var req updateWidgetReq
	if err := rest.Decode(r, &req); err != nil {
		return errs.From(err)
	}

	w, err := h.core.Update(ctx, id, widget.UpdateWidget{Name: req.Name, Description: req.Description})
	if err != nil {
		return errs.From(err)
	}
	return rest.JSON(toResponse(w))
}

func (h *Handler) delete(ctx context.Context, r *http.Request) rest.Encoder {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid id %q", r.PathValue("id"))
	}

	if err := h.core.Delete(ctx, id); err != nil {
		return errs.From(err)
	}
	return nil // 204 No Content
}
