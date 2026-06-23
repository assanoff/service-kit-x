package widget

import (
	"time"

	widgetcore "github.com/assanoff/service-kit-x/core/widget"
)

type CreateWidgetReq struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateWidgetReq struct {
	Name        *string `json:"name" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// ImportWidgetsReq is a batch enqueued for asynchronous bulk import. Name is an
// optional dedup key: re-posting the same name is a no-op.
type ImportWidgetsReq struct {
	Name    string            `json:"name" validate:"max=200"`
	Widgets []CreateWidgetReq `json:"widgets" validate:"required,min=1,max=1000,dive"`
}

type ImportResponse struct {
	Scheduled bool `json:"scheduled"`
	Count     int  `json:"count"`
}

type WidgetResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(w widgetcore.Widget) WidgetResponse {
	return WidgetResponse{
		ID:          w.ID.String(),
		Name:        w.Name,
		Description: w.Description,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func toResponses(ws []widgetcore.Widget) []WidgetResponse {
	out := make([]WidgetResponse, len(ws))
	for i, w := range ws {
		out[i] = toResponse(w)
	}
	return out
}
