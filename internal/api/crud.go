package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
)

// pathID reads a positive int32 path param, writing a 400 and returning ok=false when it is bad.
func (s *Server) pathID(w http.ResponseWriter, r *http.Request, name string) (int32, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 32)
	if err != nil || id <= 0 {
		s.writeError(w, http.StatusBadRequest, "invalid "+name, err)
		return 0, false
	}
	return int32(id), true
}

// updateGame edits a games display fields.
func (s *Server) updateGame(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[dto.UpdateGameRequest](s, w, r)
	if !ok {
		return
	}

	if req.Monetization != "f2p" && req.Monetization != "premium" {
		s.writeError(w, http.StatusBadRequest, "monetization must be 'f2p' or 'premium'", nil)
		return
	}

	refs, err := marshalRefs(req.SourceRefs)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "encode source_refs", err)
		return
	}

	row, err := s.q.UpdateGameDisplay(r.Context(), db.UpdateGameDisplayParams{
		ExternalGameID: id,
		Name:           req.Name,
		Short:          req.Short,
		Monetization:   req.Monetization,
		DisplayColor:   req.Color,
		SourceRefs:     refs,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "update game", err)
		return
	}

	s.writeJSON(w, http.StatusOK, dto.Game{
		ID:           dto.GameID(row.ExternalGameID),
		Name:         row.Name,
		Short:        row.Short,
		Monetization: row.Monetization,
		Color:        row.DisplayColor,
		SourceRefs:   refsMap(row.SourceRefs),
	})
}

// deleteGame removes a game. it is frozen (409) while any artifact references it, since artifacts have
// no FK and would orphan.
func (s *Server) deleteGame(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	n, err := s.q.CountArtifactsForGame(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "count artifacts", err)
		return
	}
	if n > 0 {
		s.writeError(w, http.StatusConflict, "game has artifacts and cannot be deleted", nil)
		return
	}

	if err := s.q.DeleteGameDisplay(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete game", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getPrompt returns one prompts full body so the edit form can prefill the template and system prompt.
func (s *Server) getPrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	p, err := s.q.GetPrompt(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "prompt not found", err)
		return
	}

	system := ""
	if p.SystemPrompt != nil {
		system = *p.SystemPrompt
	}

	s.writeJSON(w, http.StatusOK, dto.PromptDetail{
		ID:           int(p.ID),
		Name:         p.Name,
		Version:      int(p.Version),
		Modality:     p.Modality,
		SystemPrompt: system,
		Template:     p.Template,
	})
}

// createPrompt inserts a new prompt version.
func (s *Server) createPrompt(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[dto.PromptRequest](s, w, r)
	if !ok {
		return
	}
	if !validPrompt(s, w, req) {
		return
	}

	id, err := s.q.CreatePrompt(r.Context(), db.CreatePromptParams{
		Name:         req.Name,
		Version:      req.Version,
		Modality:     req.Modality,
		SystemPrompt: ptrIfSet(req.SystemPrompt),
		Template:     req.Template,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "create prompt", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, dto.Prompt{
		ID:       int(id),
		Name:     req.Name,
		Version:  int(req.Version),
		Modality: req.Modality,
	})
}

// updatePrompt edits a prompt. it is frozen (409) while any run uses it, since editing would change the
// content under runs that already ran against it.
func (s *Server) updatePrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[dto.PromptRequest](s, w, r)
	if !ok {
		return
	}
	if !validPrompt(s, w, req) {
		return
	}

	used, err := s.q.CountRunsForPrompt(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "count runs", err)
		return
	}
	if used > 0 {
		s.writeError(w, http.StatusConflict, "prompt is used by a run and cannot be edited", nil)
		return
	}

	if err := s.q.UpdatePrompt(r.Context(), db.UpdatePromptParams{
		ID:           id,
		Name:         req.Name,
		Version:      req.Version,
		Modality:     req.Modality,
		SystemPrompt: ptrIfSet(req.SystemPrompt),
		Template:     req.Template,
	}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "update prompt", err)
		return
	}

	s.writeJSON(w, http.StatusOK, dto.Prompt{
		ID:       int(id),
		Name:     req.Name,
		Version:  int(req.Version),
		Modality: req.Modality,
	})
}

// deletePrompt removes a prompt. frozen (409) while any run uses it.
func (s *Server) deletePrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	used, err := s.q.CountRunsForPrompt(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "count runs", err)
		return
	}
	if used > 0 {
		s.writeError(w, http.StatusConflict, "prompt is used by a run and cannot be deleted", nil)
		return
	}

	if err := s.q.DeletePrompt(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete prompt", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updateAnnotator edits an annotators label.
func (s *Server) updateAnnotator(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[dto.UpdateAnnotatorRequest](s, w, r)
	if !ok {
		return
	}
	if req.Label == "" {
		s.writeError(w, http.StatusBadRequest, "label is required", nil)
		return
	}

	if err := s.q.UpdateAnnotatorLabel(r.Context(), db.UpdateAnnotatorLabelParams{
		ID:    id,
		Label: req.Label,
	}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "update annotator", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteAnnotator removes an annotator. frozen (409) while any run or annotation references it.
func (s *Server) deleteAnnotator(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	refs, err := s.q.CountAnnotatorRefs(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "count refs", err)
		return
	}
	if refs > 0 {
		s.writeError(w, http.StatusConflict, "annotator is referenced by a run and cannot be deleted", nil)
		return
	}

	if err := s.q.DeleteAnnotator(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete annotator", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// runImpact reports the blast radius of deleting a run.
func (s *Server) runImpact(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	row, err := s.q.RunImpact(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "run impact", err)
		return
	}
	s.writeJSON(w, http.StatusOK, dto.ImpactResult{
		Annotations:   int(row.Annotations),
		Samples:       int(row.Samples),
		Adjudications: int(row.Adjudications),
	})
}

// populationImpact reports the blast radius of deleting a population.
func (s *Server) populationImpact(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	row, err := s.q.PopulationImpact(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "population impact", err)
		return
	}
	s.writeJSON(w, http.StatusOK, dto.ImpactResult{
		Individuals: int(row.Individuals),
		Runs:        int(row.Runs),
		Annotations: int(row.Annotations),
		Samples:     int(row.Samples),
	})
}

// deleteRun cascade-deletes a run and everything that hangs off it.
func (s *Server) deleteRun(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	if err := cascadeRun(ctx, db.New(tx), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete run", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deletePopulation cascade-deletes a population: each run over it, then its individuals, then the row.
func (s *Server) deletePopulation(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	runIDs, err := q.ListRunIDsByPopulation(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "list runs", err)
		return
	}
	for _, runID := range runIDs {
		if err := cascadeRun(ctx, q, runID); err != nil {
			s.writeError(w, http.StatusInternalServerError, "delete run", err)
			return
		}
	}

	if err := q.DeleteIndividualsByPopulation(ctx, id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete individuals", err)
		return
	}
	if err := q.DeletePopulation(ctx, id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete population", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// cascadeRun deletes one run and its children in FK order: annotation_patterns, annotations,
// adjudication samples, adjudications, then the run (run_annotators cascade on the run delete).
func cascadeRun(ctx context.Context, q *db.Queries, runID int32) error {
	if err := q.DeleteAnnotationPatternsByRun(ctx, runID); err != nil {
		return err
	}
	if err := q.DeleteAnnotationsByRun(ctx, runID); err != nil {
		return err
	}
	if err := q.DeleteAdjudicationSamplesByRun(ctx, runID); err != nil {
		return err
	}
	if err := q.DeleteAdjudicationsByRun(ctx, runID); err != nil {
		return err
	}
	return q.DeleteRun(ctx, runID)
}

// validPrompt checks the shared prompt fields, writing a 400 and returning false when bad.
func validPrompt(s *Server, w http.ResponseWriter, req dto.PromptRequest) bool {
	if req.Name == "" || req.Template == "" {
		s.writeError(w, http.StatusBadRequest, "name and template are required", nil)
		return false
	}
	if req.Version <= 0 {
		s.writeError(w, http.StatusBadRequest, "version must be a positive integer", nil)
		return false
	}
	if req.Modality != "text" && req.Modality != "image" && req.Modality != "multimodal" {
		s.writeError(w, http.StatusBadRequest, "modality must be 'text', 'image' or 'multimodal'", nil)
		return false
	}
	return true
}

func ptrIfSet(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// marshalRefs encodes a source->handle map as jsonb, with a nil map becoming '{}'.
func marshalRefs(m map[string]string) ([]byte, error) {
	if m == nil {
		m = map[string]string{}
	}
	return json.Marshal(m)
}

// refsMap decodes a source_refs jsonb blob into a map; a null/invalid blob reads as empty.
func refsMap(b []byte) map[string]string {
	m := map[string]string{}
	_ = json.Unmarshal(b, &m)
	return m
}
