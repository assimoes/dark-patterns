package api

import (
	"net/http"
	"strconv"

	"github.com/assimoes/dsr/internal/annotate"
	"github.com/assimoes/dsr/internal/api/dto"
)

// AnalysisRow is one (review, model, pattern) cell of a run
// present = true: the model detected a given pattern in the review
// preset = false: the model did not detect a given pattern in the review (true negative)
type AnalysisRow struct {
	IndividualID   int64  `json:"individualId"`
	ExternalGameID int32  `json:"externalGameId"`
	ModelSlug      string `json:"modelSlug"`
	Code           string `json:"code"`
	Present        bool   `json:"present"`
}

// AnalysisGoldRow is one adjudicated cell by the human annotator.
// Pass can be either blind or open. An open pass is one where the annotator sees the panel verdict
// a blind pass means the annotator did not see the panel verdict
type AnalysisGoldRow struct {
	IndividualID int64  `json:"individualId"`
	Code         string `json:"code"`
	Pass         string `json:"pass"`
	FinalLabel   bool   `json:"finalLabel"`
	Direction    string `json:"direction"`
}

// AnalysisStatusRow is the raw status of one (review, model) cell — completed or a parse_error — with the
// finish reason, so the notebook can measure data quality before any metric.
type AnalysisStatusRow struct {
	IndividualID int64  `json:"individualId"`
	ModelSlug    string `json:"modelSlug"`
	Status       string `json:"status"`
	FinishReason string `json:"finishReason"`
}

// AnalysisRunMeta is the run configuration settings. Useful to know what changed from run to run.
type AnalysisRunMeta struct {
	RunID           int      `json:"runId"`
	Label           string   `json:"label"`
	PopulationID    int      `json:"populationId"`
	TaxonomyVersion int      `json:"taxonomyVersion"`
	Prompt          string   `json:"prompt"`
	GameContext     bool     `json:"gameContext"`
	Temperature     float64  `json:"temperature"`
	Panel           []string `json:"panel"`
}

// AnalysisExport is everything the notebook needs for one run: the run configuration, the long-format panel
// annotations, the population adjudicated gold (both passes), and the raw panel status. The notebook
// calls this once per run.
type AnalysisExport struct {
	Run         AnalysisRunMeta     `json:"run"`
	Annotations []AnalysisRow       `json:"annotations"`
	Gold        []AnalysisGoldRow   `json:"gold"`
	Status      []AnalysisStatusRow `json:"status"`
}

func (s *Server) analysisExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	run, err := s.q.GetRun(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "get run", err)
		return
	}

	version := int32(1)
	if run.TaxonomyVersion != nil {
		version = *run.TaxonomyVersion
	}

	rows, err := s.q.AnalysisAnnotations(ctx, run.ID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load annotations", err)
		return
	}

	annotations := make([]AnalysisRow, 0, len(rows))
	for _, row := range rows {
		annotations = append(annotations, AnalysisRow{
			IndividualID:   row.IndividualID,
			ExternalGameID: row.ExternalGameID,
			ModelSlug:      row.ModelSlug,
			Code:           row.Code,
			Present:        row.Present,
		})
	}

	gold := make([]AnalysisGoldRow, 0)
	if goldRun, err := s.q.GetGoldRunForPopulation(ctx, run.PopulationID); err == nil {
		grows, err := s.q.AnalysisGold(ctx, goldRun)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "load gold", err)
			return
		}
		for _, row := range grows {
			gold = append(gold, AnalysisGoldRow{
				IndividualID: row.IndividualID,
				Code:         row.Code,
				Pass:         row.Pass,
				FinalLabel:   row.FinalLabel,
				Direction:    row.Direction,
			})
		}
	}

	srows, err := s.q.AnalysisStatus(ctx, run.ID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load status", err)
		return
	}
	status := make([]AnalysisStatusRow, 0, len(srows))
	for _, row := range srows {
		status = append(status, AnalysisStatusRow{
			IndividualID: row.IndividualID,
			ModelSlug:    row.ModelSlug,
			Status:       row.Status,
			FinishReason: row.FinishReason,
		})
	}

	prompt, err := s.q.GetPrompt(ctx, run.PromptID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load prompt", err)
		return
	}
	panel, _ := s.panelSlugs(ctx, run.ID)

	s.writeJSON(w, http.StatusOK, AnalysisExport{
		Run: AnalysisRunMeta{
			RunID:           int(run.ID),
			Label:           dto.RunLabel(run.RunType, run.ID),
			PopulationID:    int(run.PopulationID),
			TaxonomyVersion: int(version),
			Prompt:          prompt.Name + " v" + strconv.Itoa(int(prompt.Version)),
			GameContext:     annotate.UsesGameContext(prompt.Template),
			Temperature:     numericFloat(run.Temperature),
			Panel:           panel,
		},
		Annotations: annotations,
		Gold:        gold,
		Status:      status,
	})
}
