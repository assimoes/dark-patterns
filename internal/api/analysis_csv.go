package api

import (
	"archive/zip"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
)

// analysisExportCSV streams the analysis datasets as a zip of CSVs, for an external notebook (e.g. Google
// Colab) that cannot reach the database: download the zip, upload it, read the CSVs with pandas.
// ?run=<id> exports one run; ?run=all exports every llm_panel run. Every row carries a run_id column so an
// all-runs export stays separable per run.
func (s *Server) analysisExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runParam := r.URL.Query().Get("run")
	if runParam == "" {
		s.writeError(w, http.StatusBadRequest, "missing run query parameter (a run id or 'all')", nil)
		return
	}

	var runIDs []int32
	if runParam == "all" {
		runs, err := s.q.ListRuns(ctx)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "list runs", err)
			return
		}
		for _, run := range runs {
			if run.RunType == "llm_panel" {
				runIDs = append(runIDs, run.ID)
			}
		}
	} else {
		id, err := strconv.ParseInt(runParam, 10, 32)
		if err != nil || id <= 0 {
			s.writeError(w, http.StatusBadRequest, "invalid run id", err)
			return
		}
		runIDs = []int32{int32(id)}
	}

	// collect everything before writing the body, so a query error still becomes a clean HTTP error
	// (once the zip starts streaming the status line is already sent).
	exports := make([]AnalysisExport, 0, len(runIDs))
	for _, id := range runIDs {
		export, err := s.collectAnalysis(ctx, id)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "collect analysis", err)
			return
		}
		exports = append(exports, export)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="analysis-export.zip"`)

	zw := zip.NewWriter(w)
	defer zw.Close()

	if err := writeAnalysisCSVs(zw, exports); err != nil {
		// headers are already sent; the body is partial. log and let the client notice the truncated zip.
		s.logger.Error("write analysis zip", "err", err)
	}
}

// writeAnalysisCSVs writes the four CSV members into the zip, unioned across exports with a run_id column.
func writeAnalysisCSVs(zw *zip.Writer, exports []AnalysisExport) error {
	if err := writeCSV(zw, "annotations.csv",
		[]string{"run_id", "individual_id", "external_game_id", "model_slug", "code", "present"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, a := range e.Annotations {
					if err := emit([]string{
						rid,
						strconv.FormatInt(a.IndividualID, 10),
						strconv.Itoa(int(a.ExternalGameID)),
						a.ModelSlug,
						a.Code,
						strconv.FormatBool(a.Present),
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "gold.csv",
		[]string{"run_id", "individual_id", "code", "pass", "final_label", "direction"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, g := range e.Gold {
					if err := emit([]string{
						rid,
						strconv.FormatInt(g.IndividualID, 10),
						g.Code,
						g.Pass,
						strconv.FormatBool(g.FinalLabel),
						g.Direction,
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "status.csv",
		[]string{"run_id", "individual_id", "model_slug", "status", "finish_reason"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, st := range e.Status {
					if err := emit([]string{
						rid,
						strconv.FormatInt(st.IndividualID, 10),
						st.ModelSlug,
						st.Status,
						st.FinishReason,
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "meta.csv",
		[]string{"run_id", "label", "population_id", "taxonomy_version", "prompt", "game_context", "temperature", "panel"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				if err := emit([]string{
					strconv.Itoa(e.Run.RunID),
					e.Run.Label,
					strconv.Itoa(e.Run.PopulationID),
					strconv.Itoa(e.Run.TaxonomyVersion),
					e.Run.Prompt,
					strconv.FormatBool(e.Run.GameContext),
					strconv.FormatFloat(e.Run.Temperature, 'f', -1, 64),
					strings.Join(e.Run.Panel, ";"),
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "failures.csv",
		[]string{"run_id", "individual_id", "model_slug", "status", "finish_reason", "completion_tokens", "raw_response"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, f := range e.Failures {
					if err := emit([]string{
						rid,
						strconv.FormatInt(f.IndividualID, 10),
						f.ModelSlug,
						f.Status,
						f.FinishReason,
						f.CompletionTokens,
						f.RawResponse,
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "pattern_distribution.csv",
		[]string{"run_id", "external_game_id", "game_name", "code", "reviews"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, d := range e.PatternDistribution {
					if err := emit([]string{
						rid,
						strconv.Itoa(int(d.ExternalGameID)),
						d.GameName,
						d.Code,
						strconv.Itoa(int(d.Reviews)),
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	if err := writeCSV(zw, "sample.csv",
		[]string{"run_id", "sample_id", "panel_run_id", "individual_id", "stratum", "selection_prob"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, s := range e.SampleStrata {
					if err := emit([]string{
						rid,
						strconv.FormatInt(s.SampleID, 10),
						strconv.Itoa(int(s.PanelRunID)),
						strconv.FormatInt(s.IndividualID, 10),
						s.Stratum,
						strconv.FormatFloat(s.SelectionProb, 'f', -1, 64),
					}); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
		return err
	}

	return writeCSV(zw, "seed.csv",
		[]string{"run_id", "individual_id", "code", "model_slug", "present"},
		func(emit func([]string) error) error {
			for _, e := range exports {
				rid := strconv.Itoa(e.Run.RunID)
				for _, v := range e.SeedVotes {
					if err := emit([]string{
						rid,
						strconv.FormatInt(v.IndividualID, 10),
						v.Code,
						v.ModelSlug,
						strconv.FormatBool(v.Present),
					}); err != nil {
						return err
					}
				}
			}
			return nil
		})
}

// writeCSV creates one named file in the zip, writes the header row, then lets fill emit the data rows.
func writeCSV(zw *zip.Writer, name string, header []string, fill func(emit func([]string) error) error) error {
	f, err := zw.Create(name)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(f)
	if err := cw.Write(header); err != nil {
		return err
	}
	if err := fill(cw.Write); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}
