// Command run creates an annotation run row pinning a population, prompt, taxonomy version and the
// panel of annotators. validates all of it exists first so a run never points at nothing.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/assimoes/dsr/internal/db"
	"github.com/assimoes/dsr/internal/run"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var (
		populationID = flag.Int("population", 0, "population id to annotate (required)")
		promptID     = flag.Int("prompt", 0, "prompt id the panel runs (required)")
		taxVersion   = flag.Int("taxonomy-version", 1, "taxonomy generation the run pins")
		runType      = flag.String("type", "llm_panel", "run type: llm_panel | gold")
		temperature  = flag.Float64("temperature", 0, "sampling temperature")
		annotators   = flag.String("annotators", "", "annotators to be used in this run")
	)

	flag.Parse()

	if *populationID == 0 || *promptID == 0 {
		logger.Error("both -population and -prompt are required")
		os.Exit(2)
	}

	if *runType != "llm_panel" && *runType != "gold" {
		logger.Error("invalid -type (want llm_panel | gold)", "type", *runType)
		os.Exit(2)
	}

	annotatorIDs, err := parseAnnotatorsIDs(*annotators)
	if err != nil {
		logger.Error("invalid -annotators", "err", err)
		os.Exit(2)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL not set")
		os.Exit(2)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect to postgres failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	if err := run.Validate(ctx, q, annotatorIDs, int32(*populationID), int32(*promptID), int32(*taxVersion)); err != nil {
		logger.Error("validation failed", "err", err)
		os.Exit(1)
	}

	var temp pgtype.Numeric
	if err := temp.Scan(fmt.Sprintf("%g", *temperature)); err != nil {
		logger.Error("encode temperature", "err", err)
		os.Exit(1)
	}

	tv := int32(*taxVersion)
	runID, err := q.CreateRun(ctx, db.CreateRunParams{
		RunType:         *runType,
		PopulationID:    int32(*populationID),
		PromptID:        int32(*promptID),
		Temperature:     temp,
		TopP:            pgtype.Numeric{},
		Params:          nil,
		TaxonomyVersion: &tv,
		AnnotatorIds:    annotatorIDs,
	})

	if err != nil {
		logger.Error("create run", "err", err)
		os.Exit(1)
	}

	logger.Info("run created",
		"run_id", runID,
		"type", *runType,
		"population", *populationID,
		"prompt", *promptID,
		"taxonomy_version", *taxVersion,
		"temperature", *temperature,
	)
}

// parseAnnotatorsIDs splits the comma list of annotator ids into int32s, skipping blanks. empty
// string means no annotators.
func parseAnnotatorsIDs(csv string) ([]int32, error) {
	csv = strings.TrimSpace(csv)

	if csv == "" {
		return nil, nil
	}

	var ids []int32
	for _, f := range strings.Split(csv, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("not an integer annotator id: %q", f)
		}

		ids = append(ids, int32(n))
	}

	return ids, nil
}
