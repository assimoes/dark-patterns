package ingest

import "github.com/jackc/pgx/v5/pgtype"

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric

	if s == "" {
		return n, nil
	}

	err := n.Scan(s)
	return n, err
}
