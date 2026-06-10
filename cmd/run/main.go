package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/assimoes/dsr/internal/annotate"
	"github.com/assimoes/dsr/internal/db"
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

	// validate before inserting

	if err := validate(ctx, q, annotatorIDs, int32(*populationID), int32(*promptID), int32(*taxVersion)); err != nil {
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

func validate(ctx context.Context, q *db.Queries, annotatorIDs []int32, populationID, promptID, taxVersion int32) error {

	// ensure annotators exist and are llms

	if len(annotatorIDs) > 0 {
		rows, err := q.ListAnnotatorsByIDs(ctx, annotatorIDs)
		if err != nil {
			return fmt.Errorf("loading annotators %v: %w", annotatorIDs, err)
		}

		found := make(map[int32]db.Annotator, len(rows))
		for _, a := range rows {
			found[a.ID] = a
		}

		for _, id := range annotatorIDs {
			a, ok := found[id]

			if !ok {
				return fmt.Errorf("annotator %d does not exist", id)
			}

			if a.Kind != "llm" || a.ModelID == nil {
				return fmt.Errorf("annotator %d is not an llm annotator", id)
			}
		}
	}

	// ensure population exists and has individuals
	n, err := q.CountIndividuals(ctx, populationID)
	if err != nil {
		return fmt.Errorf("counting individuals for population %d: %w", populationID, err)
	}

	if n == 0 {
		return fmt.Errorf("population %d has no individuals (did you curate it? use the curate cli)", populationID)
	}

	// ensure prompt exists
	prompt, err := q.GetPrompt(ctx, promptID)
	if err != nil {
		return fmt.Errorf("loading prompt %d: %w", promptID, err)
	}

	if prompt.Modality != "text" && prompt.Modality != "image" {
		return fmt.Errorf("prompt %d has unsupported modality %q", promptID, prompt.Modality)
	}

	// ensure the taxonomy has codes
	codes, err := q.ListMesoPatternsByVersion(ctx, taxVersion)
	if err != nil {
		return fmt.Errorf("loading taxonomy version %d: %w", taxVersion, err)
	}

	if len(codes) == 0 {
		return fmt.Errorf("taxonomy version %d has no codes", taxVersion)
	}

	_ = annotate.LoadTaxonomy

	return nil
}

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
