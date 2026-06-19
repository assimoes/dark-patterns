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

// promptData backs both the system and user templates. taxonomy block feeds simple
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

// Loader fetches one item and renders it into a model Input. one per modality.
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
		Nonce:    nonce,
	}); err != nil {
		return Input{}, err
	}

	return Input{
		System: rc.System,
		User:   b.String(),
	}, nil
}

// ImageLoader pulls the screenshot for an individual and renders the multimodal Input: the prompt text
// plus the image attached for the model to look at.
type ImageLoader struct{}

// Modality is "image".
func (ImageLoader) Modality() string {
	return "image"
}

// Load fetches the image uri, renders the user template (any OCR text rides along in Content), and
// returns an Input carrying the image so the annotator sends it as an image_url part.
func (ImageLoader) Load(ctx context.Context, q *db.Queries,
	rc RenderCtx, individualID int64) (Input, error) {

	row, err := q.GetImageForIndividual(ctx, individualID)
	if err != nil {
		return Input{}, err
	}

	user, err := renderUser(rc, row.OcrText)
	if err != nil {
		return Input{}, err
	}

	mime := ""
	if row.MimeType != nil {
		mime = *row.MimeType
	}

	return Input{
		System: rc.System,
		User:   user,
		Images: []Image{{URL: row.ImageUri, MIMEType: mime}},
	}, nil
}

// MultimodalLoader pulls an artifact that has both a text body and an image, and fills one Input with
// both: the body as the user text and the image attached.
type MultimodalLoader struct{}

// Modality is "multimodal".
func (MultimodalLoader) Modality() string {
	return "multimodal"
}

// Load reads both channels for the individual and returns an Input carrying the body and the image.
func (MultimodalLoader) Load(ctx context.Context, q *db.Queries,
	rc RenderCtx, individualID int64) (Input, error) {

	row, err := q.GetMultimodalForIndividual(ctx, individualID)
	if err != nil {
		return Input{}, err
	}

	user, err := renderUser(rc, row.Body)
	if err != nil {
		return Input{}, err
	}

	mime := ""
	if row.MimeType != nil {
		mime = *row.MimeType
	}

	return Input{
		System: rc.System,
		User:   user,
		Images: []Image{{URL: row.ImageUri, MIMEType: mime}},
	}, nil
}

func newNonce() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

// renderUser runs the per-item user template with a fresh nonce. shared by the image and multimodal
// loaders so each Load stays short. (textLoader renders its own, since it also passes language/voted_up.)
func renderUser(rc RenderCtx, content string) (string, error) {
	nonce, err := newNonce()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	if err := rc.Template.Execute(&b, promptData{
		Taxonomy: rc.Taxonomy,
		Content:  content,
		Nonce:    nonce,
	}); err != nil {
		return "", err
	}

	return b.String(), nil
}
