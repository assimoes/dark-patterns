package dto

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func GameID(id int32) string {
	return strconv.FormatInt(int64(id), 10)
}

func RunLabel(runType string, id int32) string {
	return fmt.Sprintf("%s #%s", runType, strconv.FormatInt(int64(id), 10))
}

func RFC3339(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}

	return ts.Time.UTC().Format(time.RFC3339)
}

func PopulationLabel(id int32, desc *string) string {
	if desc != nil && *desc != "" {
		return *desc
	}

	return fmt.Sprintf("Population #%s", strconv.FormatInt(int64(id), 10))
}

func IntPtr(v *int32) *int {
	if v == nil {
		return nil
	}

	n := int(*v)
	return &n
}
