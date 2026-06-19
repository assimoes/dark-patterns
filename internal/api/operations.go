package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/assimoes/dsr/internal/annotate"
	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/run"
	"github.com/assimoes/dsr/internal/scrape"
	"github.com/jackc/pgx/v5/pgtype"
)

// createGame registers a game so it shows on the dashboard and can be scraped.
func (s *Server) createGame(w http.ResponseWriter, r *http.Request) {

	req, ok := decodeJSON[dto.AddGameRequest](s, w, r)
	if !ok {
		return
	}

	if req.ExternalGameID <= 0 {
		s.writeError(w, http.StatusBadRequest, "external_game_id must be a positive integer", nil)
		return
	}

	name := req.Name
	if name == "" {
		name = fmt.Sprintf("Game %d", req.ExternalGameID)
	}

	short := req.Short
	if short == "" {
		short = strconv.FormatInt(int64(req.ExternalGameID), 10)
	}

	monetization := req.Monetization
	if monetization == "" {
		monetization = "f2p"
	}
	if monetization != "f2p" && monetization != "premium" {
		s.writeError(w, http.StatusBadRequest, "monetization must be 'f2p' or 'premium'", nil)
		return
	}

	color := req.Color
	if color == "" {
		color = "#64748b"
	}

	if _, err := s.q.InsertGameDisplay(r.Context(), db.InsertGameDisplayParams{
		ExternalGameID: req.ExternalGameID,
		Name:           name,
		Short:          short,
		Monetization:   monetization,
		DisplayColor:   color,
	}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "register game", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, dto.Game{
		ID:           dto.GameID(req.ExternalGameID),
		Name:         name,
		Short:        short,
		Monetization: monetization,
		Color:        color,
	})
}

// createAnnotator inserts an annotator.
func (s *Server) createAnnotator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := decodeJSON[dto.AddAnnotatorRequest](s, w, r)
	if !ok {
		return
	}

	if req.Label == "" {
		s.writeError(w, http.StatusBadRequest, "label is required", nil)
		return
	}

	switch req.Kind {
	case "human":
		id, err := s.q.CreateAnnotator(ctx, db.CreateAnnotatorParams{
			Kind:    "human",
			ModelID: nil,
			Label:   req.Label,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "create human annotator", err)
			return
		}
		s.writeJSON(w, http.StatusCreated, dto.AnnotatorResponse{ID: id, Kind: "human", Label: req.Label})

	case "llm":
		if req.Family == "" || req.Slug == "" || req.Name == "" {
			s.writeError(w, http.StatusBadRequest, "llm annotators require family, slug and name", nil)
			return
		}
		modalities := req.Modalities
		if len(modalities) == 0 {
			modalities = []string{"text"}
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "begin tx", err)
			return
		}
		defer tx.Rollback(ctx)

		q := db.New(tx)

		modelID, err := q.UpsertModel(ctx, db.UpsertModelParams{
			Family:     req.Family,
			Slug:       req.Slug,
			Name:       req.Name,
			Modalities: modalities,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "upsert model", err)
			return
		}

		id, err := q.CreateAnnotator(ctx, db.CreateAnnotatorParams{
			Kind:    "llm",
			ModelID: &modelID,
			Label:   req.Label,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "create llm annotator", err)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			s.writeError(w, http.StatusInternalServerError, "commit tx", err)
			return
		}

		s.writeJSON(w, http.StatusCreated, dto.AnnotatorResponse{ID: id, Kind: "llm", Label: req.Label, ModelID: &modelID})

	default:
		s.writeError(w, http.StatusBadRequest, "kind must be 'human' or 'llm'", nil)
	}
}

// createPopulation creates a population, freezes a stratified sample into it, and reports the size.
func (s *Server) createPopulation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := decodeJSON[dto.CreatePopulationRequest](s, w, r)
	if !ok {
		return
	}

	modality := req.Modality
	if modality == "" {
		modality = "text"
	}
	if modality != "text" && modality != "image" && modality != "multimodal" {
		s.writeError(w, http.StatusBadRequest, "modality must be 'text', 'image' or 'multimodal'", nil)
		return
	}

	minHours := req.MinHoursPlayed
	if minHours == 0 {
		minHours = 1
	}
	perGame := req.PerGameCap
	if perGame == 0 {
		perGame = 50
	}

	cut := time.Now()
	if req.ArtifactsCutoff != "" {
		t, err := time.Parse(time.RFC3339, req.ArtifactsCutoff)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "artifacts_cutoff must be RFC3339", err)
			return
		}
		cut = t
	}

	critJSON, err := json.Marshal(dto.OpsCriteria{
		Modality:        "text",
		MinHoursPlayed:  int32(minHours),
		PerGameCap:      perGame,
		ArtifactsCutoff: cut,
		GameIDs:         req.GameIDs,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "marshal criteria", err)
		return
	}

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	popID, err := q.CreatePopulation(ctx, db.CreatePopulationParams{
		Modality:        "text",
		Description:     desc,
		Criteria:        critJSON,
		ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "create population", err)
		return
	}

	var inserted int64

	switch modality {
	case "image":
		inserted, err = q.FreezeImagePopulation(ctx, db.FreezeImagePopulationParams{
			PopulationID:    popID,
			ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
			PerGameCap:      int32(perGame),
		})
	case "multimodal":
		inserted, err = q.FreezeMultimodalPopulation(ctx, db.FreezeMultimodalPopulationParams{
			PopulationID:    popID,
			ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
			PerGameCap:      int32(perGame),
		})
	default:
		inserted, err = q.FreezeStratifiedPopulation(ctx, db.FreezeStratifiedPopulationParams{
			PopulationID:    popID,
			ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
			MinHoursPlayed:  int32(minHours),
			PerGameCap:      int32(perGame),
			GameIds:         req.GameIDs,
		})
	}

	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "freeze population", err)
		return
	}

	total, err := q.CountIndividuals(ctx, popID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "count individuals", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, dto.CreatePopulationResponse{
		PopulationID:        popID,
		InsertedIndividuals: inserted,
		TotalIndividuals:    total,
	})
}

// createRun validates the inputs exactly as the run CLI does, then inserts the run row.
func (s *Server) createRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := decodeJSON[dto.CreateRunRequest](s, w, r)
	if !ok {
		return
	}

	if req.PopulationID == 0 || req.PromptID == 0 {
		s.writeError(w, http.StatusBadRequest, "population_id and prompt_id are required", nil)
		return
	}

	runType := req.RunType
	if runType == "" {
		runType = "llm_panel"
	}
	if runType != "llm_panel" && runType != "gold" {
		s.writeError(w, http.StatusBadRequest, "run_type must be 'llm_panel' or 'gold'", nil)
		return
	}

	taxVersion := req.TaxonomyVersion
	if taxVersion == 0 {
		taxVersion = 1
	}

	if err := run.Validate(ctx, s.q, req.AnnotatorIDs, req.PopulationID, req.PromptID, taxVersion); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var temp pgtype.Numeric
	if err := temp.Scan(fmt.Sprintf("%g", req.Temperature)); err != nil {
		s.writeError(w, http.StatusBadRequest, "encode temperature", err)
		return
	}

	tv := taxVersion
	runID, err := s.q.CreateRun(ctx, db.CreateRunParams{
		RunType:         runType,
		PopulationID:    req.PopulationID,
		PromptID:        req.PromptID,
		Temperature:     temp,
		TopP:            pgtype.Numeric{},
		Params:          nil,
		TaxonomyVersion: &tv,
		AnnotatorIds:    req.AnnotatorIDs,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "create run", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, dto.CreateRunResponse{
		RunID:           runID,
		RunType:         runType,
		PopulationID:    req.PopulationID,
		PromptID:        req.PromptID,
		TaxonomyVersion: taxVersion,
		AnnotatorIDs:    req.AnnotatorIDs,
	})
}

// createScrape enqueues the first scrape job for a game, exactly like the steam CLI
func (s *Server) createScrape(w http.ResponseWriter, r *http.Request) {

	req, ok := decodeJSON[dto.ScrapeRequest](s, w, r)
	if !ok {
		return
	}

	app, err := strconv.ParseInt(req.App, 10, 32)
	if err != nil || app <= 0 {
		s.writeError(w, http.StatusBadRequest, "app must be a positive Steam app id", err)
		return
	}

	filter := req.Filter
	if filter == "" {
		filter = "recent"
	}
	if filter != "recent" && filter != "updated" {
		s.writeError(w, http.StatusBadRequest, "filter must be 'recent' or 'updated'", nil)
		return
	}

	lang := req.Lang
	if lang == "" {
		lang = "english"
	}

	max := req.Max
	if max == 0 {
		max = 500
	}
	if max < 0 {
		s.writeError(w, http.StatusBadRequest, "max must be >= 0", nil)
		return
	}

	if err := scrape.Enqueue(r.Context(), s.riverClient, s.pool, int32(app), filter, lang, max); err != nil {
		s.writeError(w, http.StatusInternalServerError, "enqueue scrape", err)
		return
	}

	s.writeJSON(w, http.StatusAccepted, map[string]any{
		"enqueued": true,
		"game_id":  int32(app),
		"filter":   filter,
		"language": lang,
		"max":      max,
	})
}

// enqueueAnnotations freezes the runs panel and enqueues one annotation job per (individual, annotator),
// same as annotate enqueue. POST /api/runs/{runID}/annotations.
func (s *Server) enqueueAnnotations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runID, err := strconv.ParseInt(r.PathValue("runID"), 10, 32)
	if err != nil || runID <= 0 {
		s.writeError(w, http.StatusBadRequest, "invalid run id", err)
		return
	}

	registry, err := s.panelRegistry(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "build panel registry", err)
		return
	}

	n, err := annotate.Enqueue(ctx, s.riverClient, s.pool, int32(runID), registry)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "enqueue annotations", err)
		return
	}

	s.writeJSON(w, http.StatusAccepted, map[string]any{
		"enqueued": n,
		"run_id":   int32(runID),
	})
}

// panelRegistry builds the slug->Annotator map for the active llm panel
func (s *Server) panelRegistry(ctx context.Context) (map[string]annotate.Annotator, error) {
	anns, err := s.q.ListLLMAnnotators(ctx)
	if err != nil {
		return nil, err
	}

	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		return nil, errors.New("OPENROUTER_API_KEY not set")
	}

	httpClient := &http.Client{Timeout: 90 * time.Second}

	registry := make(map[string]annotate.Annotator, len(anns))
	for _, a := range anns {
		registry[a.Slug] = annotate.NewOpenRouterAnnotator(
			key,
			a.Slug,
			annotate.WithHTTPClient(httpClient),
			annotate.WithClientVersion("openrouter/v1"),
			annotate.WithAttribution("https://github.com/assimoes/dark-patterns", "design science research artifact"),
		)
	}

	return registry, nil
}

func (s *Server) createImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := decodeJSON[dto.UploadImagesRequest](s, w, r)
	if !ok {
		return
	}

	if req.GameID <= 0 {
		s.writeError(w, http.StatusBadRequest, "game_id must be a positive integer", nil)
		return
	}

	if len(req.Images) == 0 {
		s.writeError(w, http.StatusBadRequest, "at least one image is required", nil)
		return
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	now := time.Now()
	ids := make([]int64, 0, len(req.Images))

	for _, img := range req.Images {
		raw, err := base64.StdEncoding.DecodeString(img.Data)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "data must be base64", err)
			return
		}

		mime := img.MimeType
		if mime == "" {
			mime = "image/png"
		}

		sum := sha256.Sum256(raw)
		dataURI := fmt.Sprintf("data:%s;base64,%s", mime, img.Data)
		sourceID := fmt.Sprintf("upload-%d-%x", req.GameID, sum[:8])

		artifactID, err := q.UpsertArtifact(ctx, db.UpsertArtifactParams{
			Modality:       "image",
			Source:         "upload",
			SourceID:       &sourceID,
			ContentHash:    sum[:],
			ScrapedAt:      pgtype.Timestamptz{Time: now, Valid: true},
			ExternalGameID: req.GameID,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "store artifact", err)
			return
		}

		var desc *string
		if img.Description != "" {
			desc = &img.Description
		}

		mimeCopy := mime
		if err := q.UpsertImageDetail(ctx, db.UpsertImageDetailParams{
			ArtifactID:  artifactID,
			ImageUri:    dataURI,
			MimeType:    &mimeCopy,
			Description: desc,
		}); err != nil {
			s.writeError(w, http.StatusInternalServerError, "store image detail", err)
			return
		}

		ids = append(ids, artifactID)
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, dto.UploadImagesResponse{
		Uploaded: len(ids),
		GameID:   req.GameID,
		ImageIDs: ids,
	})

}
