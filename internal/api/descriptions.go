package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/research"
	"github.com/jackc/pgx/v5"
)

// listGameDescriptions returns a games description version history, newest first, for the review console.
func (s *Server) listGameDescriptions(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	rows, err := s.q.ListGameDescriptionsForGame(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load descriptions", err)
		return
	}

	out := make([]dto.GameDescription, 0, len(rows))
	for _, d := range rows {
		out = append(out, dto.NewGameDescription(d))
	}
	s.writeJSON(w, http.StatusOK, out)
}

// getGameDescription returns one description version.
func (s *Server) getGameDescription(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	d, err := s.q.GetGameDescription(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "description not found", err)
		return
	}
	s.writeJSON(w, http.StatusOK, dto.NewGameDescription(d))
}

// updateGameDescription saves a reviewers edits to a draft. the rendered text and valence flags are
// re-derived server-side from the edited profile, so the injected text and the neutrality check always
// match what was approved. only drafts are editable.
func (s *Server) updateGameDescription(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	profile, ok := decodeJSON[research.Profile](s, w, r)
	if !ok {
		return
	}

	rendered := research.Render(profile)
	flags := research.ScanValence(rendered)

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "encode profile", err)
		return
	}
	sourcesJSON, err := json.Marshal(profile.Sources)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "encode sources", err)
		return
	}
	flagsJSON, err := json.Marshal(flags)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "encode valence flags", err)
		return
	}

	d, err := s.q.UpdateDraftDescription(r.Context(), db.UpdateDraftDescriptionParams{
		ID:           id,
		Profile:      profileJSON,
		RenderedText: rendered,
		Sources:      sourcesJSON,
		ValenceFlags: flagsJSON,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		s.writeError(w, http.StatusConflict, "only a draft can be edited", nil)
		return
	}
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "update description", err)
		return
	}
	s.writeJSON(w, http.StatusOK, dto.NewGameDescription(d))
}

// approveGameDescription freezes a draft as the approved version: it supersedes the prior approved one
// and promotes this draft, in one transaction. blocked while any valence flag is unresolved.
func (s *Server) approveGameDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	draft, err := s.q.GetGameDescription(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "description not found", err)
		return
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	if err := q.SupersedePriorApproved(ctx, draft.ExternalGameID); err != nil {
		s.writeError(w, http.StatusInternalServerError, "supersede prior approved", err)
		return
	}

	approver := s.auditor
	approved, err := q.ApproveDescription(ctx, db.ApproveDescriptionParams{ID: id, ApprovedBy: &approver})
	if errors.Is(err, pgx.ErrNoRows) {
		s.writeError(w, http.StatusConflict, "cannot approve: not a draft, or it has unresolved valence flags", nil)
		return
	}
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "approve description", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit", err)
		return
	}
	s.writeJSON(w, http.StatusOK, dto.NewGameDescription(approved))
}

// researchGameDescription re-runs the research for a game, producing a fresh draft to review.
func (s *Server) researchGameDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	game, err := s.q.GetGameDisplay(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "game not found", err)
		return
	}

	var body struct {
		DisambiguationHint string `json:"disambiguation_hint"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := research.Enqueue(ctx, s.riverClient, id, game.Name, body.DisambiguationHint); err != nil {
		s.writeError(w, http.StatusInternalServerError, "enqueue research", err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, map[string]any{"enqueued": true, "game_id": id})
}
