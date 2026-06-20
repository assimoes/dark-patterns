package research

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// ResearchGameArgs is one onboarding research job: research one game and store the result as a draft
// description for human review.
type ResearchGameArgs struct {
	ExternalGameID     int32  `json:"external_game_id"`
	GameName           string `json:"game_name"`
	DisambiguationHint string `json:"disambiguation_hint"`
}

// Kind is the River job kind.
func (ResearchGameArgs) Kind() string { return "research_game" }

// InsertOpts routes to the research queue and caps retries, since each attempt is an expensive web
// search.
func (ResearchGameArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: "research", MaxAttempts: 3}
}

// Worker runs the research call and persists the draft.
type Worker struct {
	river.WorkerDefaults[ResearchGameArgs]
	pool   *pgxpool.Pool
	client *Client
	logger *slog.Logger
}

// NewWorker wires the research worker. nil logger falls back to slog.Default.
func NewWorker(pool *pgxpool.Pool, client *Client, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{pool: pool, client: client, logger: logger}
}

// Work researches the game, validates the result, and stores a draft. a failed call retries until the
// last attempt, then records an error draft so the reviewer sees it rather than the job vanishing. a
// result that does not parse or validate is stored as an invalid draft for the reviewer to fix.
func (w *Worker) Work(ctx context.Context, job *river.Job[ResearchGameArgs]) error {
	a := job.Args
	q := db.New(w.pool)

	version, err := q.NextDescriptionVersion(ctx, a.ExternalGameID)
	if err != nil {
		return err
	}

	raw, modelUsed, err := w.client.Research(ctx, a.GameName, a.DisambiguationHint)
	if err != nil {
		if job.Attempt < job.MaxAttempts {
			return err
		}
		w.logger.Error("research call failed", "game", a.ExternalGameID, "err", err)
		return w.persist(ctx, q, a.ExternalGameID, version, draftError, Profile{}, "", w.client.Model(), err.Error())
	}

	var profile Profile
	if perr := json.Unmarshal([]byte(raw), &profile); perr != nil {
		return w.persistInvalid(ctx, q, a.ExternalGameID, version, modelUsed, raw, "response is not valid JSON: "+perr.Error())
	}
	if verr := profile.Validate(); verr != nil {
		return w.persistInvalid(ctx, q, a.ExternalGameID, version, modelUsed, raw, verr.Error())
	}

	rendered := Render(profile)
	flags := ScanValence(rendered)
	w.logger.Info("research draft", "game", a.ExternalGameID, "version", version, "valence_flags", len(flags))

	return w.persist(ctx, q, a.ExternalGameID, version, draftOK, profile, rendered, modelUsed, "", withSources(profile.Sources), withFlags(flags))
}

const (
	draftOK    = "draft"
	draftError = "error"
	draftBad   = "invalid"
)

// persistInvalid stores an unparseable/invalid result, keeping the raw output as the profile when it is
// at least valid JSON, so the reviewer has something to fix.
func (w *Worker) persistInvalid(ctx context.Context, q *db.Queries, gameID, version int32, modelUsed, raw, msg string) error {
	var profile Profile
	if json.Valid([]byte(raw)) {
		_ = json.Unmarshal([]byte(raw), &profile)
	}
	return w.persist(ctx, q, gameID, version, draftBad, profile, "", modelUsed, msg)
}

type persistOpt func(*db.InsertGameDescriptionParams)

func withSources(sources []string) persistOpt {
	return func(p *db.InsertGameDescriptionParams) {
		if b, err := json.Marshal(sources); err == nil && sources != nil {
			p.Sources = b
		}
	}
}

func withFlags(flags []Flag) persistOpt {
	return func(p *db.InsertGameDescriptionParams) {
		if b, err := json.Marshal(flags); err == nil {
			p.ValenceFlags = b
		}
	}
}

func (w *Worker) persist(ctx context.Context, q *db.Queries, gameID, version int32,
	status string, profile Profile, rendered, modelUsed, errMsg string, opts ...persistOpt) error {

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	params := db.InsertGameDescriptionParams{
		ExternalGameID: gameID,
		Version:        version,
		Status:         status,
		Profile:        profileJSON,
		RenderedText:   rendered,
		ResearchModel:  modelUsed,
		Sources:        []byte("[]"),
		ValenceFlags:   []byte("[]"),
		Error:          errMsg,
	}
	for _, o := range opts {
		o(&params)
	}

	_, err = q.InsertGameDescription(ctx, params)
	return err
}
