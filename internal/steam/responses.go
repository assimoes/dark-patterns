package steam

import "encoding/json"

type Author struct {
	SteamID              string `json:"steamid"`
	NumGamesOwned        int    `json:"num_games_owned"`
	NumReviews           int    `json:"num_reviews"`
	PlaytimeForever      int    `json:"playtime_forever"`
	PlaytimeLastTwoWeeks int    `json:"playtime_last_two_weeks"`
	LastPlayed           int64  `json:"last_played"`
}

type Review struct {
	RecommendationID         string     `json:"recommendationid"`
	Author                   Author     `json:"author"`
	Language                 string     `json:"language"`
	Review                   string     `json:"review"`
	TimestampCreated         int64      `json:"timestamp_created"`
	TimestampUpdated         int64      `json:"timestamp_updated"`
	VotedUp                  bool       `json:"voted_up"`
	VotesUp                  int        `json:"votes_up"`
	VotesFunny               int        `json:"votes_funny"`
	WeightedVotedScore       flexString `json:"weighted_vote_score"`
	CommentCount             int        `json:"comment_count"`
	SteamPurchase            bool       `json:"steam_purchase"`
	ReceivedForFree          bool       `json:"received_for_free"`
	Refunded                 bool       `json:"refunded"`
	WrittenDuringEarlyAccess bool       `json:"written_during_early_access"`
}

type QuerySummary struct {
	NumReviews      int    `json:"num_reviews"`
	ReviewScore     int    `json:"review_score"`
	ReviewScoreDesc string `json:"review_score_desc"`
	TotalPositive   int    `json:"total_positive"`
	TotalNegative   int    `json:"total_negative"`
	TotalReviews    int    `json:"total_reviews"`
}

type ReviewResponse struct {
	Success      int          `json:"success"`
	QuerySummary QuerySummary `json:"query_summary"`
	Cursor       string       `json:"cursor"`
	Reviews      []Review     `json:"reviews"`
}

// The steam api is inconsistent with the weighted_vote_score field. Sometimes it returns "0.5" and other times 0.5 or "null".
// This is why we have to make this "hack"
type flexString string

func (s *flexString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = ""
		return nil
	}

	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}

		*s = flexString(str)
		return nil
	}

	*s = flexString(b)
	return nil
}
