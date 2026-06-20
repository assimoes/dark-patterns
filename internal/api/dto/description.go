package dto

import (
	"encoding/json"

	"github.com/assimoes/dsr/internal/db"
)

// GameDescription is one version of a games neutral description, for the review console. profile,
// sources, and valenceFlags pass through as raw json the frontend types; renderedText is the exact text
// that gets injected and hashed once approved.
type GameDescription struct {
	ID             int             `json:"id"`
	ExternalGameID string          `json:"externalGameId"`
	Version        int             `json:"version"`
	Status         string          `json:"status"`
	Profile        json.RawMessage `json:"profile"`
	RenderedText   string          `json:"renderedText"`
	ResearchModel  string          `json:"researchModel"`
	Sources        json.RawMessage `json:"sources"`
	ValenceFlags   json.RawMessage `json:"valenceFlags"`
	Error          string          `json:"error"`
	CreatedAt      string          `json:"createdAt"`
	ApprovedAt     string          `json:"approvedAt"`
}

func NewGameDescription(d db.GameDescription) GameDescription {
	return GameDescription{
		ID:             int(d.ID),
		ExternalGameID: GameID(d.ExternalGameID),
		Version:        int(d.Version),
		Status:         d.Status,
		Profile:        d.Profile,
		RenderedText:   d.RenderedText,
		ResearchModel:  d.ResearchModel,
		Sources:        d.Sources,
		ValenceFlags:   d.ValenceFlags,
		Error:          d.Error,
		CreatedAt:      RFC3339(d.CreatedAt),
		ApprovedAt:     RFC3339(d.ApprovedAt),
	}
}
