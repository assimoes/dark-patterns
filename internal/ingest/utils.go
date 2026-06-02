package ingest

import (
	"crypto/sha256"

	"github.com/assimoes/dsr/internal/steam"
	"github.com/jackc/pgx/v5/pgtype"
)

func hash(r steam.Review) []byte {
	h := sha256.New()
	h.Write([]byte(r.RecommendationID))
	h.Write([]byte{0})
	h.Write([]byte(r.Review))

	return h.Sum(nil)
}

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric

	if s == "" {
		return n, nil
	}

	err := n.Scan(s)
	return n, err
}

func ptr[T any](v T) *T { return &v }
