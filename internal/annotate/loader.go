package annotate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"text/template"

	"github.com/assimoes/dsr/internal/db"
)

// RenderCtx is the per-run prompt material. System is pre-rendered; Template is the per-item user message.
type RenderCtx struct {
	System   string
	Template *template.Template
	Taxonomy string
}

// promptData backs both the system and user templates. Taxonomy block feeds simple
// prompts; HighLevels/Patterns feed the richer ones; the rest carry the per-item review.
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

// Loader fetches one item and renders it into a model Input. One per modality.
type Loader interface {
	Modality() string
	Load(ctx context.Context, q *db.Queries, rc RenderCtx, individualID int64) (Input, error)
}

// TextLoader pulls the text review for an individual and renders the text-only Input.
type TextLoader struct{}

// Modality is "text".
func (TextLoader) Modality() string {
	return "text"
}

// Load fetches the review body, renders the user template with a fresh nonce, returns the Input.
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
