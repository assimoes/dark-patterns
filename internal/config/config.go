// Package config parses the steam scraper flags and env into one Config.
package config

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

// Config is everything the steam enqueue command needs, half from flags half from env.
type Config struct {
	DatabaseURL string
	AppID       string
	GameID      int32
	MaxReviews  int
	Language    string
	Filter      string
	PageDelay   time.Duration
}

// Load parses args for the flags and reads DATABASE_URL via getenv, then validates.
// getenv is injected so tests dont touch the real environment.
func Load(args []string, getenv func(string) string) (Config, error) {
	fs := flag.NewFlagSet("steam", flag.ContinueOnError)

	var c Config
	fs.StringVar(&c.AppID, "app", "", "steam app id")
	fs.IntVar(&c.MaxReviews, "max", 500, "max reviews to pull (0 = all)")
	fs.StringVar(&c.Language, "lang", "english", "review language. default is english")
	fs.StringVar(&c.Filter, "filter", "recent", "recent or updated")
	fs.DurationVar(&c.PageDelay, "delay", 500*time.Millisecond, "polite delay between pages")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	c.DatabaseURL = getenv("DATABASE_URL")

	return c, c.validate()
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}
	if c.AppID == "" {
		return fmt.Errorf("missing required -app flag (steam app id)")
	}

	id, err := strconv.Atoi(c.AppID)
	if err != nil {
		return fmt.Errorf("invalid app id %q: %w", c.AppID, err)
	}

	c.GameID = int32(id)

	if c.MaxReviews < 0 {
		return fmt.Errorf("-max must be >= 0, got %d", c.MaxReviews)
	}
	if c.Filter != "recent" && c.Filter != "updated" {
		return fmt.Errorf("-filter must be recent or updated, got %q", c.Filter)
	}

	return nil
}
