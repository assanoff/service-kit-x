package widgetapi

import (
	"time"

	"github.com/assanoff/service-kit-x/core/widget"
)

type createWidgetReq struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type updateWidgetReq struct {
	Name        *string `json:"name" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// importWidgetsReq is a batch enqueued for asynchronous bulk import. Name is an
// optional dedup key: re-posting the same name is a no-op.
type importWidgetsReq struct {
	Name    string            `json:"name" validate:"max=200"`
	Widgets []createWidgetReq `json:"widgets" validate:"required,min=1,max=1000,dive"`
}

type importResponse struct {
	Scheduled bool `json:"scheduled"`
	Count     int  `json:"count"`
}

type widgetResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(w widget.Widget) widgetResponse {
	return widgetResponse{
		ID:          w.ID.String(),
		Name:        w.Name,
		Description: w.Description,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func toResponses(ws []widget.Widget) []widgetResponse {
	out := make([]widgetResponse, len(ws))
	for i, w := range ws {
		out[i] = toResponse(w)
	}
	return out
}
