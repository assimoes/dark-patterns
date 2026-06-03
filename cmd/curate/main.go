package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type criteria struct {
	Modality        string    `json:"modality"`
	MinHoursPlayed  int32     `json:"min_hours_played"`
	PerGameCap      int       `json:"per_game_cap"`
	ArtifactsCutoff time.Time `json:"artifacts_cutoff"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var (
		desc     = flag.String("desc", "", "human description of this population")
		minHours = flag.Int("min-hours", 1, "minimum hours_played")
		perGame  = flag.Int("per-game", 50, "max individuals per game (stratification cap)")
		cutoff   = flag.String("cutoff", "", "artifacts_cutoff RFC3339; empty = now()")
	)

	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL not set")
		os.Exit(2)
	}

	cut := time.Now()

	if *cutoff != "" {
		t, err := time.Parse(time.RFC3339, *cutoff)
		if err != nil {
			logger.Error("bar -cutoff (want RFC3339)", "err", err)
			os.Exit(2)
		}

		cut = t
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	crit := criteria{
		Modality:        "text",
		MinHoursPlayed:  int32(*minHours),
		PerGameCap:      *perGame,
		ArtifactsCutoff: cut,
	}

	critJSON, err := json.Marshal(crit)

	if err != nil {
		logger.Error("marshal criteria", "err", err)
		os.Exit(1)
	}

	popID, err := q.CreatePopulation(ctx, db.CreatePopulationParams{
		Modality:        "text",
		Description:     desc,
		Criteria:        critJSON,
		ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
	})
	if err != nil {
		logger.Error("create population", "err", err)
		os.Exit(1)
	}

	logger.Info("created population", "id", popID, "cutoff", cut.Format(time.RFC3339))

	inserted, err := q.FreezeStratifiedPopulation(ctx, db.FreezeStratifiedPopulationParams{
		PopulationID:    popID,
		ArtifactsCutoff: pgtype.Timestamptz{Time: cut, Valid: true},
		MinHoursPlayed:  int32(*minHours),
		PerGameCap:      int32(*perGame),
	})
	if err != nil {
		logger.Error("freeze population", "err", err)
		os.Exit(1)
	}

	total, err := q.CountIndividuals(ctx, popID)
	if err != nil {
		logger.Error("count individuals", "err", err)
		os.Exit(1)
	}

	logger.Info(
		"population frozen",
		"population_id", popID,
		"inserted", inserted,
		"individuals", total,
		"cutoff", cut.Format(time.RFC3339))

}
