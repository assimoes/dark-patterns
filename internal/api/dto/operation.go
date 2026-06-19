package dto

import "time"

// UploadImageItem one screenshot in an upload
type UploadImageItem struct {
	Data        string `json:"data"`
	MimeType    string `json:"mimeType"`
	Description string `json:"description"`
}

// UploadImagesRequest body of POST /api/images
type UploadImagesRequest struct {
	GameID int32             `json:"game_id"`
	Images []UploadImageItem `json:"images"`
}

// UploadImagesResponse 201 body
type UploadImagesResponse struct {
	Uploaded int     `json:"uploaded"`
	GameID   int32   `json:"game_id"`
	ImageIDs []int64 `json:"image_ids"`
}

// ScrapeRequest is the body of POST /api/scrapes, same args as steam enqueue. App is the steam app id as a string.
type ScrapeRequest struct {
	App    string `json:"app"`
	Filter string `json:"filter"`
	Lang   string `json:"lang"`
	Max    int    `json:"max"`
}

// CreateRunRequest is the body of POST /api/runs, same inputs cmd/run takes.
type CreateRunRequest struct {
	PopulationID    int32   `json:"population_id"`
	PromptID        int32   `json:"prompt_id"`
	TaxonomyVersion int32   `json:"taxonomy_version"`
	RunType         string  `json:"run_type"`
	Temperature     float64 `json:"temperature"`
	AnnotatorIDs    []int32 `json:"annotator_ids"`
}

// CreateRunResponse is the 201 body of POST /api/runs, echoes the resolved run back.
type CreateRunResponse struct {
	RunID           int32   `json:"run_id"`
	RunType         string  `json:"run_type"`
	PopulationID    int32   `json:"population_id"`
	PromptID        int32   `json:"prompt_id"`
	TaxonomyVersion int32   `json:"taxonomy_version"`
	AnnotatorIDs    []int32 `json:"annotator_ids"`
}

// CreatePopulationRequest is the body of POST /api/populations, same knobs the curate CLI takes.
// gameIds is empty for all games, or the picked external ids to restrict the freeze to.
type CreatePopulationRequest struct {
	Description     string  `json:"description"`
	MinHoursPlayed  int     `json:"min_hours_played"`
	PerGameCap      int     `json:"per_game_cap"`
	ArtifactsCutoff string  `json:"artifacts_cutoff"`
	GameIDs         []int32 `json:"game_ids"`
	Modality        string  `json:"modality"`
}

// CreatePopulationResponse is the 201 body: the new id, how many rows the freeze inserted, and the running total.
type CreatePopulationResponse struct {
	PopulationID        int32 `json:"population_id"`
	InsertedIndividuals int64 `json:"inserted_individuals"`
	TotalIndividuals    int64 `json:"total_individuals"`
}

// OpsCriteria is the criteria JSON the curate CLI stores on the population row.
type OpsCriteria struct {
	Modality        string    `json:"modality"`
	MinHoursPlayed  int32     `json:"min_hours_played"`
	PerGameCap      int       `json:"per_game_cap"`
	ArtifactsCutoff time.Time `json:"artifacts_cutoff"`
	GameIDs         []int32   `json:"game_ids,omitempty"`
}

// AddAnnotatorRequest is the body of POST /api/annotators. the family/slug/name/modalities fields only
// matter for kind=llm, theyre the model row we upsert before linking the annotator to it.
type AddAnnotatorRequest struct {
	Kind       string   `json:"kind"`
	Label      string   `json:"label"`
	Family     string   `json:"family"`
	Slug       string   `json:"slug"`
	Name       string   `json:"name"`
	Modalities []string `json:"modalities"`
}

// AnnotatorResponse is the 201 body of POST /api/annotators. ModelID is null for humans.
type AnnotatorResponse struct {
	ID      int32  `json:"id"`
	Kind    string `json:"kind"`
	Label   string `json:"label"`
	ModelID *int32 `json:"model_id"`
}

// AddGameRequest is the body of POST /api/games, the fields that become a game_display row.
type AddGameRequest struct {
	ExternalGameID int32  `json:"external_game_id"`
	Name           string `json:"name"`
	Short          string `json:"short"`
	Monetization   string `json:"monetization"`
	Color          string `json:"color"`
}
