package api

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// populationLabel is a populations description, or "Population #<id>" when its unset.
func populationLabel(id int32, desc *string) string {
	if desc != nil && *desc != "" {
		return *desc
	}

	return "Population #" + strconv.FormatInt(int64(id), 10)
}

// gameID renders an external_game_id as the string id the frontend uses.
func gameID(id int32) string {
	return strconv.FormatInt(int64(id), 10)
}

// parseGameID turns a {gameId} path value into an external_game_id.
func parseGameID(s string) (int32, error) {
	return parseInt32(s)
}

// parseInt32 parses a base-10 int32 from a path value.
func parseInt32(s string) (int32, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(n), nil
}

// parseLabel maps "present"/"absent" to a final label; ok is false otherwise.
func parseLabel(s string) (label, ok bool) {
	switch s {
	case "present":
		return true, true
	case "absent":
		return false, true
	default:
		return false, false
	}
}

// runLabel is "<run_type> #<id>"; the runs table has no name column.
func runLabel(runType string, id int32) string {
	return runType + " #" + strconv.FormatInt(int64(id), 10)
}

// rfc3339 renders a nullable timestamp as RFC3339, empty when null.
func rfc3339(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}

	return ts.Time.UTC().Format(time.RFC3339)
}
