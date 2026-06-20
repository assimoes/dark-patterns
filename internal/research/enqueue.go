package research

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// Enqueue kicks off the research for a game by inserting one job. called when a game is added (and again
// for an explicit re-research).
func Enqueue(ctx context.Context, client *river.Client[pgx.Tx], gameID int32, gameName, hint string) error {
	_, err := client.Insert(ctx, ResearchGameArgs{
		ExternalGameID:     gameID,
		GameName:           gameName,
		DisambiguationHint: hint,
	}, nil)
	return err
}
