package steam

// FetchOpts is what to ask steam for: which app, language, filter and how many.
type FetchOpts struct {
	AppID      string
	Language   string
	Filter     string
	MaxReviews int
	MaxPerPage int
}

// normalize fills the blanks so a zero-value FetchOpts still works: language all, recent filter,
// page size clamped to the steam max.
func (opts *FetchOpts) normalize() {
	if opts.Language == "" {
		opts.Language = "all"
	}

	if opts.Filter == "" {
		opts.Filter = "recent"
	}

	if opts.MaxPerPage <= 0 || opts.MaxPerPage > maxPerPage {
		opts.MaxPerPage = maxPerPage
	}
}

// retryable wraps an error we want to back off and try again on, vs a hard fail we give up on.
type retryable struct {
	err error
}

func (e retryable) Error() string {
	return e.err.Error()
}

func (e retryable) Unwrap() error {
	return e.err
}
