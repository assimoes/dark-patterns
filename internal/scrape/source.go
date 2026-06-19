package scrape

import (
	"context"

	"github.com/assimoes/dsr/internal/ingest"
)

// Source fetches one page of scraped items for a target and names its queue. a new source is a new
// Source in the registry; the worker does not change. params carry source-specific knobs (steam: filter,
// language) so the job args stay source-agnostic.
type Source interface {
	Name() string
	Fetch(ctx context.Context, target, cursor string, params map[string]string) (items []ingest.Item, next string, err error)
}
