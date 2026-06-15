package api

import (
	"strconv"
)

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
