package annotate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"text/template"

	"github.com/assimoes/dsr/internal/db"
)

// RenderCtx is the per-run prompt material, shared across modalities
// handed on every run. System is already rendered; Template is the per-item
// user message.
type RenderCtx struct {
	System   string
	Template *template.Template
	Taxonomy string
}

// promptData is the single context both the system and user templates render
// against. A flat taxonomy block feeds the simpler prompts; the structured
// HighLevels/Patterns feed the richer ones. Per-item fields carry the review.
type promptData struct {
	Taxonomy        string
	Content         string
	HighLevels      []HighLevelView
	Patterns        []PatternView
	TaxonomyVersion string
	PromptVersion   string
	Language        string
	VotedUp         bool
	Nonce           string
}

// Loader fetches one item's content and renders it into the model Input
// One implementation per modality.
type Loader interface {
	Modality() string
	Load(ctx context.Context, q *db.Queries, rc RenderCtx, individualID int64) (Input, error)
}

type TextLoader struct{}

func (TextLoader) Modality() string {
	return "text"
}

func (TextLoader) Load(ctx context.Context, q *db.Queries,
	rc RenderCtx, individualID int64) (Input, error) {

	row, err := q.GetTextReviewForIndividual(ctx, individualID)
	if err != nil {
		return Input{}, err
	}

	nonce, err := newNonce()
	if err != nil {
		return Input{}, err
	}

	var b strings.Builder
	if err := rc.Template.Execute(&b, promptData{
		Taxonomy: rc.Taxonomy,
		Content:  row.Body,
		Language: row.Lang,
		VotedUp:  row.VotedUp,
		Nonce:    nonce,
	}); err != nil {
		return Input{}, err
	}

	return Input{
		System: rc.System,
		User:   b.String(),
	}, nil
}

func newNonce() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
