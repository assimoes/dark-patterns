package api

import (
	"context"
	"fmt"
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
// AnalysisFailureRow is one annotation that did not complete: its model, status, finish reason, completion
// token count, and the raw text the model returned that failed to parse — the input for failure-mode analysis.
type AnalysisFailureRow struct {
	IndividualID     int64  `json:"individualId"`
	ModelSlug        string `json:"modelSlug"`
	Status           string `json:"status"`
	FinishReason     string `json:"finishReason"`
	CompletionTokens string `json:"completionTokens"`
	RawResponse      string `json:"rawResponse"`
}

// AnalysisPatternDistRow is one (game, pattern) cell of the run's detection profile: how many reviews a
// majority of the completed panel flagged for that meso code in that game.
type AnalysisPatternDistRow struct {
	ExternalGameID int32  `json:"externalGameId"`
	GameName       string `json:"gameName"`
	Code           string `json:"code"`
	Reviews        int32  `json:"reviews"`
}

// AnalysisSampleRow is one (sample, review) of the stratified adjudication sample behind the gold run: the
// stratum it was drawn from and the inverse-probability selection weight — so the notebook can show the gold
// is a stratified subset, not the corpus.
type AnalysisSampleRow struct {
	SampleID      int64   `json:"sampleId"`
	PanelRunID    int32   `json:"panelRunId"`
	IndividualID  int64   `json:"individualId"`
	Stratum       string  `json:"stratum"`
	SelectionProb float64 `json:"selectionProb"`
}

// AnalysisSeedRow is one per-model vote frozen in an adjudication's seed (the panel that was on screen at
// decision time). Lets the notebook verify the live annotations of the adjudicated run match the seed.
type AnalysisSeedRow struct {
	IndividualID int64  `json:"individualId"`
	Code         string `json:"code"`
	ModelSlug    string `json:"modelSlug"`
	Present      bool   `json:"present"`
}

type AnalysisExport struct {
	Run                 AnalysisRunMeta          `json:"run"`
	Annotations         []AnalysisRow            `json:"annotations"`
	Gold                []AnalysisGoldRow        `json:"gold"`
	Status              []AnalysisStatusRow      `json:"status"`
	Failures            []AnalysisFailureRow     `json:"failures"`
	PatternDistribution []AnalysisPatternDistRow `json:"patternDistribution"`
	SampleStrata        []AnalysisSampleRow      `json:"sampleStrata"`
	SeedVotes           []AnalysisSeedRow        `json:"seedVotes"`
}

func (s *Server) analysisExport(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r, "id")
	if !ok {
		return
	}

	export, err := s.collectAnalysis(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "analysis export", err)
		return
	}

	s.writeJSON(w, http.StatusOK, export)
}

// collectAnalysis assembles the full analysis dataset for one run: the run config, the long-format panel
// annotations, the population's adjudicated gold (both passes), and the raw panel status. Shared by the JSON
// per-run handler and the CSV/zip export.
func (s *Server) collectAnalysis(ctx context.Context, runID int32) (AnalysisExport, error) {
	run, err := s.q.GetRun(ctx, runID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("get run %d: %w", runID, err)
	}

	version := int32(1)
	if run.TaxonomyVersion != nil {
		version = *run.TaxonomyVersion
	}

	rows, err := s.q.AnalysisAnnotations(ctx, run.ID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("load annotations: %w", err)
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
	sampleStrata := make([]AnalysisSampleRow, 0)
	seedVotes := make([]AnalysisSeedRow, 0)
	if goldRun, err := s.q.GetGoldRunForPopulation(ctx, run.PopulationID); err == nil {
		grows, err := s.q.AnalysisGold(ctx, goldRun)
		if err != nil {
			return AnalysisExport{}, fmt.Errorf("load gold: %w", err)
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

		strataRows, err := s.q.AnalysisSampleStrata(ctx, goldRun)
		if err != nil {
			return AnalysisExport{}, fmt.Errorf("load sample strata: %w", err)
		}
		for _, row := range strataRows {
			sampleStrata = append(sampleStrata, AnalysisSampleRow{
				SampleID:      row.SampleID,
				PanelRunID:    row.PanelRunID,
				IndividualID:  row.IndividualID,
				Stratum:       row.Stratum,
				SelectionProb: row.SelectionProb,
			})
		}

		seedRows, err := s.q.AnalysisSeedVotes(ctx, goldRun)
		if err != nil {
			return AnalysisExport{}, fmt.Errorf("load seed votes: %w", err)
		}
		for _, row := range seedRows {
			seedVotes = append(seedVotes, AnalysisSeedRow{
				IndividualID: row.IndividualID,
				Code:         row.Code,
				ModelSlug:    row.ModelSlug,
				Present:      row.Present,
			})
		}
	}

	srows, err := s.q.AnalysisStatus(ctx, run.ID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("load status: %w", err)
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

	frows, err := s.q.AnalysisFailures(ctx, run.ID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("load failures: %w", err)
	}
	failures := make([]AnalysisFailureRow, 0, len(frows))
	for _, row := range frows {
		failures = append(failures, AnalysisFailureRow{
			IndividualID:     row.IndividualID,
			ModelSlug:        row.ModelSlug,
			Status:           row.Status,
			FinishReason:     row.FinishReason,
			CompletionTokens: row.CompletionTokens,
			RawResponse:      row.RawResponse,
		})
	}

	drows, err := s.q.RunPatternDistribution(ctx, run.ID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("load pattern distribution: %w", err)
	}
	dist := make([]AnalysisPatternDistRow, 0, len(drows))
	for _, row := range drows {
		dist = append(dist, AnalysisPatternDistRow{
			ExternalGameID: row.ExternalGameID,
			GameName:       row.GameName,
			Code:           row.Code,
			Reviews:        row.Reviews,
		})
	}

	prompt, err := s.q.GetPrompt(ctx, run.PromptID)
	if err != nil {
		return AnalysisExport{}, fmt.Errorf("load prompt: %w", err)
	}
	panel, _ := s.panelSlugs(ctx, run.ID)

	return AnalysisExport{
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
		Annotations:         annotations,
		Gold:                gold,
		Status:              status,
		Failures:            failures,
		PatternDistribution: dist,
		SampleStrata:        sampleStrata,
		SeedVotes:           seedVotes,
	}, nil
}
