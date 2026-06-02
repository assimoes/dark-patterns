package steam

type FetchOpts struct {
	AppID      string
	Language   string
	Filter     string
	MaxReviews int
	MaxPerPage int
}

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

type retryable struct {
	err error
}

func (e retryable) Error() string {
	return e.err.Error()
}

func (e retryable) Unwrap() error {
	return e.err
}
