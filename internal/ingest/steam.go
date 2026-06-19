package ingest

import "github.com/assimoes/dsr/internal/steam"

// SteamItem maps a steam review into a normalised Item. body and lang are the text channel; the
// steam-only fields (voted_up, hours, score) ride in source_meta. playtime comes in minutes, divide
// by 60 for hours.
func SteamItem(r steam.Review) Item {
	return Item{
		SourceID: r.RecommendationID,
		Body:     r.Review,
		Lang:     r.Language,
		Meta: map[string]any{
			"voted_up":            r.VotedUp,
			"hours_played":        int32(r.Author.PlaytimeForever / 60),
			"weighted_vote_score": string(r.WeightedVotedScore),
		},
	}
}
