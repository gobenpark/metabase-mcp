package metabase

import (
	"encoding/json"
	"strconv"
	"time"
)

// FlexID handles Metabase IDs that can be int or string (e.g. "root" collection).
type FlexID struct {
	IntVal    int
	StringVal string
	IsString  bool
}

func (f *FlexID) UnmarshalJSON(data []byte) error {
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		f.IntVal = i
		f.IsString = false
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.StringVal = s
		f.IsString = true
		return nil
	}
	return nil
}

func (f FlexID) String() string {
	if f.IsString {
		return f.StringVal
	}
	return strconv.Itoa(f.IntVal)
}

// Database represents a Metabase database connection.
type Database struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Engine string `json:"engine"`
}

// Card represents a saved question in Metabase.
type Card struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Display     string `json:"display"`
	Archived    bool   `json:"archived"`
}

// CardDetail is a Card with its dataset query included (from GET /api/card/:id).
type CardDetail struct {
	Card
	DatasetQuery CardDatasetQuery `json:"dataset_query"`
}

// CardDatasetQuery wraps the query stages returned by the Metabase card API.
type CardDatasetQuery struct {
	Database int              `json:"database"`
	Type     string           `json:"type"`
	Native   *NativeQuery     `json:"native,omitempty"`
	// New MBQL stages format
	Stages []CardQueryStage `json:"stages,omitempty"`
}

// CardQueryStage is a single stage in the new MBQL query format.
type CardQueryStage struct {
	Native string `json:"native,omitempty"`
}

// Dashboard represents a Metabase dashboard.
type Dashboard struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Cards       []DashboardCard `json:"dashcards"`
	CreatorID   int             `json:"creator_id"`
	CreatedAt   time.Time       `json:"created_at"`
}

// DashboardCard represents a card within a dashboard.
type DashboardCard struct {
	ID     int  `json:"id"`
	CardID *int `json:"card_id"`
	Card   Card `json:"card"`
	Row    int  `json:"row"`
	Col    int  `json:"col"`
	SizeX  int  `json:"size_x"`
	SizeY  int  `json:"size_y"`
}

// Collection represents a Metabase collection (folder).
type Collection struct {
	ID          FlexID `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Archived    bool   `json:"archived"`
}

// SearchResult represents a Metabase search result item.
type SearchResult struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
}

// SearchResponse represents the search API response.
type SearchResponse struct {
	Data       []SearchResult `json:"data"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

// CreateCollectionRequest is the payload for creating a collection (folder).
type CreateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ParentID    *int   `json:"parent_id,omitempty"`
}

// CreateDashboardRequest is the payload for creating a dashboard.
type CreateDashboardRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	CollectionID *int   `json:"collection_id,omitempty"`
}

// AddCardToDashboardRequest is the payload for adding a card to a dashboard.
type AddCardToDashboardRequest struct {
	CardID int `json:"card_id"`
	Row    int `json:"row"`
	Col    int `json:"col"`
	SizeX  int `json:"size_x"`
	SizeY  int `json:"size_y"`
}

// CreateCardRequest is the payload for creating a saved question (card).
type CreateCardRequest struct {
	Name                  string         `json:"name"`
	Type                  string         `json:"type,omitempty"`
	Description           string         `json:"description,omitempty"`
	Display               string         `json:"display"`
	CollectionID          *int           `json:"collection_id,omitempty"`
	DatasetQuery          DatasetQuery   `json:"dataset_query"`
	VisualizationSettings map[string]any `json:"visualization_settings"`
}

// UpdateCardRequest is the payload for updating a saved question (card).
// All fields are optional — only non-zero values are sent.
type UpdateCardRequest struct {
	Name                  string         `json:"name,omitempty"`
	Type                  string         `json:"type,omitempty"`
	Description           string         `json:"description,omitempty"`
	Display               string         `json:"display,omitempty"`
	DatasetQuery          *DatasetQuery  `json:"dataset_query,omitempty"`
	VisualizationSettings map[string]any `json:"visualization_settings,omitempty"`
}

// DatasetQuery is the payload for executing a native query.
type DatasetQuery struct {
	Database int              `json:"database"`
	Type     string           `json:"type"`
	Native   *NativeQuery     `json:"native,omitempty"`
}

// NativeQuery represents a native (SQL) query.
type NativeQuery struct {
	Query string `json:"query"`
}

// QueryResult represents the result of a dataset query.
type QueryResult struct {
	Data   QueryData `json:"data"`
	Status string    `json:"status"`
	Error  string    `json:"error,omitempty"`
}

// QueryData holds column info and rows from a query result.
type QueryData struct {
	Columns []string        `json:"cols_names,omitempty"`
	Cols    []QueryColumn   `json:"cols"`
	Rows    [][]any `json:"rows"`
}

// QueryColumn describes a single column in the query result.
type QueryColumn struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	BaseType    string `json:"base_type"`
}
