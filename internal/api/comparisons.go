package api

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/assimoes/dsr/internal/annotate"
	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// createComparison saves a comparison of two runs. there are no constraints: the comparison runs over
// whatever reviews and models the two runs share, and shows nothing where they don't overlap.
func (s *Server) createComparison(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := decodeJSON[dto.CreateComparisonRequest](s, w, r)
	if !ok {
		return
	}
	if req.Label == "" {
		s.writeError(w, http.StatusBadRequest, "label is required", nil)
		return
	}
	if req.RunAID == req.RunBID {
		s.writeError(w, http.StatusBadRequest, "pick two different runs", nil)
		return
	}

	runA, err := s.q.GetRun(ctx, int32(req.RunAID))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "run A not found", err)
		return
	}
	runB, err := s.q.GetRun(ctx, int32(req.RunBID))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "run B not found", err)
		return
	}

	c, err := s.q.InsertComparison(ctx, db.InsertComparisonParams{
		Label:    req.Label,
		RunAID:   runA.ID,
		RunBID:   runB.ID,
		Contract: []byte("{}"),
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "save comparison", err)
		return
	}
	s.writeJSON(w, http.StatusCreated, dto.NewComparison(c))
}

// listComparisons returns the saved comparisons, newest first.
func (s *Server) listComparisons(w http.ResponseWriter, r *http.Request) {
	rows, err := s.q.ListComparisons(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load comparisons", err)
		return
	}
	out := make([]dto.Comparison, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.NewComparison(c))
	}
	s.writeJSON(w, http.StatusOK, out)
}

// deleteComparison removes a saved comparison.
func (s *Server) deleteComparison(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := s.q.DeleteComparison(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, "delete comparison", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getComparison returns the full comparison: the two runs, the attribute contract status, the per-model
// agreement, and the review totals.
func (s *Server) getComparison(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	c, err := s.q.GetComparison(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "comparison not found", err)
		return
	}

	runA, err := s.q.GetRun(ctx, c.RunAID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run A", err)
		return
	}
	runB, err := s.q.GetRun(ctx, c.RunBID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run B", err)
		return
	}
	promptA, err := s.q.GetPrompt(ctx, runA.PromptID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load prompt A", err)
		return
	}
	promptB, err := s.q.GetPrompt(ctx, runB.PromptID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load prompt B", err)
		return
	}
	panelA, _ := s.panelSlugs(ctx, runA.ID)
	panelB, _ := s.panelSlugs(ctx, runB.ID)
	attrs := attributeDiffs(runA, runB, promptA, promptB, panelA, panelB)

	agree, err := s.q.ComparisonModelAgreement(ctx, db.ComparisonModelAgreementParams{RunA: runA.ID, RunB: runB.ID})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "model agreement", err)
		return
	}
	agreement := make([]dto.ModelAgreement, 0, len(agree))
	for _, a := range agree {
		agreement = append(agreement, dto.ModelAgreement{Model: a.ModelSlug, AgreedPresent: int(a.AgreedPresent), Flips: int(a.Flips)})
	}

	rows, err := s.q.ComparisonReviewDivergence(ctx, db.ComparisonReviewDivergenceParams{RunA: runA.ID, RunB: runB.ID})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "divergence summary", err)
		return
	}
	var diverged int
	for _, row := range rows {
		if row.Flips > 0 {
			diverged++
		}
	}

	s.writeJSON(w, http.StatusOK, dto.ComparisonDetail{
		Comparison:      dto.NewComparison(c),
		RunA:            dto.RunRef{ID: int(runA.ID), Label: dto.RunLabel(runA.RunType, runA.ID), RunType: runA.RunType},
		RunB:            dto.RunRef{ID: int(runB.ID), Label: dto.RunLabel(runB.RunType, runB.ID), RunType: runB.RunType},
		Attributes:      attrs,
		ModelAgreement:  agreement,
		ReviewsTotal:    len(rows),
		ReviewsDiverged: diverged,
	})
}

// comparisonReviews returns the divergence worklist for a comparison.
func (s *Server) comparisonReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	c, err := s.q.GetComparison(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "comparison not found", err)
		return
	}

	rows, err := s.q.ComparisonReviewDivergence(ctx, db.ComparisonReviewDivergenceParams{RunA: c.RunAID, RunB: c.RunBID})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "divergence summary", err)
		return
	}
	out := make([]dto.ComparisonReviewRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.NewComparisonReviewRow(row))
	}
	s.writeJSON(w, http.StatusOK, out)
}

// panelSlugs returns the sorted model slugs of a run's frozen panel.
func (s *Server) panelSlugs(ctx context.Context, runID int32) ([]string, error) {
	annotators, err := s.q.ListRunAnnotators(ctx, runID)
	if err != nil {
		return nil, err
	}
	slugs := make([]string, 0, len(annotators))
	for _, a := range annotators {
		slugs = append(slugs, a.ModelSlug)
	}
	sort.Strings(slugs)
	return slugs, nil
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// attributeDiffs builds the per-attribute side-by-side of the two runs — informational context for
// reading the divergence, not a constraint.
func attributeDiffs(runA, runB db.Run, promptA, promptB db.Prompt, panelA, panelB []string) []dto.AttributeDiff {
	row := func(attr, a, b string, equal bool) dto.AttributeDiff {
		return dto.AttributeDiff{Attribute: attr, A: a, B: b, Equal: equal}
	}

	yesNo := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	promptLabel := func(p db.Prompt) string { return p.Name + " v" + strconv.Itoa(int(p.Version)) }

	ctxA, ctxB := annotate.UsesGameContext(promptA.Template), annotate.UsesGameContext(promptB.Template)
	tempA, tempB := numericFloat(runA.Temperature), numericFloat(runB.Temperature)

	return []dto.AttributeDiff{
		row("population", strconv.Itoa(int(runA.PopulationID)), strconv.Itoa(int(runB.PopulationID)), runA.PopulationID == runB.PopulationID),
		row("panel", strings.Join(panelA, ", "), strings.Join(panelB, ", "), equalStrings(panelA, panelB)),
		row("taxonomy", taxLabel(runA.TaxonomyVersion), taxLabel(runB.TaxonomyVersion), eqInt32Ptr(runA.TaxonomyVersion, runB.TaxonomyVersion)),
		row("prompt", promptLabel(promptA), promptLabel(promptB), runA.PromptID == runB.PromptID),
		row("temperature", strconv.FormatFloat(tempA, 'g', -1, 64), strconv.FormatFloat(tempB, 'g', -1, 64), tempA == tempB),
		row("gameContext", yesNo(ctxA), yesNo(ctxB), ctxA == ctxB),
		row("runType", runA.RunType, runB.RunType, runA.RunType == runB.RunType),
	}
}

func taxLabel(v *int32) string {
	if v == nil {
		return "—"
	}
	return "v" + strconv.Itoa(int(*v))
}

func eqInt32Ptr(a, b *int32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func numericFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN {
		return 0
	}
	v, err := n.Value()
	if err != nil {
		return 0
	}
	s, ok := v.(string)
	if !ok {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
