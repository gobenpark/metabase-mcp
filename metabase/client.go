package metabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client communicates with the Metabase REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClientFromEnv creates a Client from METABASE_URL and METABASE_API_KEY env vars.
func NewClientFromEnv() (*Client, error) {
	baseURL := strings.TrimRight(os.Getenv("METABASE_URL"), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("METABASE_URL environment variable is required")
	}
	apiKey := os.Getenv("METABASE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("METABASE_API_KEY environment variable is required")
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("metabase API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// ListDatabases returns all databases connected to Metabase.
func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	var resp struct {
		Data []Database `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/database", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ExecuteNativeQuery runs a native SQL query against the specified database.
func (c *Client) ExecuteNativeQuery(ctx context.Context, databaseID int, query string) (*QueryResult, error) {
	payload := DatasetQuery{
		Database: databaseID,
		Type:     "native",
		Native:   &NativeQuery{Query: query},
	}
	var result QueryResult
	if err := c.do(ctx, http.MethodPost, "/api/dataset", payload, &result); err != nil {
		return nil, err
	}
	if result.Status == "failed" {
		return nil, fmt.Errorf("query failed: %s", result.Error)
	}
	return &result, nil
}

// RunQuestion executes a saved question (card) by its ID.
func (c *Client) RunQuestion(ctx context.Context, cardID int) (*QueryResult, error) {
	var result QueryResult
	path := fmt.Sprintf("/api/card/%d/query", cardID)
	if err := c.do(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	if result.Status == "failed" {
		return nil, fmt.Errorf("question execution failed: %s", result.Error)
	}
	return &result, nil
}

// ListDashboards returns all dashboards.
func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	var dashboards []Dashboard
	if err := c.do(ctx, http.MethodGet, "/api/dashboard", nil, &dashboards); err != nil {
		return nil, err
	}
	return dashboards, nil
}

// GetDashboard returns a single dashboard with its cards.
func (c *Client) GetDashboard(ctx context.Context, id int) (*Dashboard, error) {
	var dashboard Dashboard
	path := fmt.Sprintf("/api/dashboard/%d", id)
	if err := c.do(ctx, http.MethodGet, path, nil, &dashboard); err != nil {
		return nil, err
	}
	return &dashboard, nil
}

// ListCollections returns all collections.
func (c *Client) ListCollections(ctx context.Context) ([]Collection, error) {
	var collections []Collection
	if err := c.do(ctx, http.MethodGet, "/api/collection", nil, &collections); err != nil {
		return nil, err
	}
	return collections, nil
}

// CreateCollection creates a new collection (folder).
func (c *Client) CreateCollection(ctx context.Context, req CreateCollectionRequest) (*Collection, error) {
	var collection Collection
	if err := c.do(ctx, http.MethodPost, "/api/collection", req, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// DeleteDashboard deletes a dashboard by ID.
func (c *Client) DeleteDashboard(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/dashboard/%d", id)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// DeleteCollection deletes a collection by ID.
func (c *Client) DeleteCollection(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/collection/%d", id)
	// Metabase archives collections rather than hard-deleting
	payload := map[string]any{"archived": true}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// MoveDashboard moves a dashboard to a different collection.
func (c *Client) MoveDashboard(ctx context.Context, dashboardID int, collectionID int) error {
	path := fmt.Sprintf("/api/dashboard/%d", dashboardID)
	payload := map[string]any{"collection_id": collectionID}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// CreateDashboard creates a new dashboard.
func (c *Client) CreateDashboard(ctx context.Context, req CreateDashboardRequest) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, http.MethodPost, "/api/dashboard", req, &dashboard); err != nil {
		return nil, err
	}
	return &dashboard, nil
}

// AddCardToDashboard adds a saved question (card) to a dashboard.
func (c *Client) AddCardToDashboard(ctx context.Context, dashboardID int, req AddCardToDashboardRequest) error {
	path := fmt.Sprintf("/api/dashboard/%d", dashboardID)

	// Metabase expects PUT with full dashcards array, so we fetch first then append.
	existing, err := c.GetDashboard(ctx, dashboardID)
	if err != nil {
		return fmt.Errorf("get dashboard: %w", err)
	}

	newCard := map[string]any{
		"id":                     -1,
		"card_id":                req.CardID,
		"row":                    req.Row,
		"col":                    req.Col,
		"size_x":                 req.SizeX,
		"size_y":                 req.SizeY,
		"parameter_mappings":     []any{},
		"visualization_settings": map[string]any{},
	}

	cards := make([]map[string]any, 0, len(existing.Cards)+1)
	for _, dc := range existing.Cards {
		cards = append(cards, map[string]any{
			"id":                     dc.ID,
			"card_id":                dc.CardID,
			"row":                    dc.Row,
			"col":                    dc.Col,
			"size_x":                 dc.SizeX,
			"size_y":                 dc.SizeY,
			"parameter_mappings":     []any{},
			"visualization_settings": map[string]any{},
		})
	}
	cards = append(cards, newCard)

	payload := map[string]any{
		"dashcards": cards,
	}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// ArchiveCard archives a saved question (card) by setting archived=true.
func (c *Client) ArchiveCard(ctx context.Context, cardID int) error {
	path := fmt.Sprintf("/api/card/%d", cardID)
	payload := map[string]any{"archived": true}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// CreateCard creates a new saved question (card).
func (c *Client) CreateCard(ctx context.Context, req CreateCardRequest) (*Card, error) {
	var card Card
	if err := c.do(ctx, http.MethodPost, "/api/card", req, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// UpdateCardDisplay changes the display (chart) type of a saved question (card).
func (c *Client) UpdateCardDisplay(ctx context.Context, cardID int, display string) error {
	path := fmt.Sprintf("/api/card/%d", cardID)
	payload := map[string]any{"display": display}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// UpdateCard updates a saved question (card). Only non-zero fields in req are applied.
func (c *Client) UpdateCard(ctx context.Context, cardID int, req UpdateCardRequest) (*Card, error) {
	path := fmt.Sprintf("/api/card/%d", cardID)
	var card Card
	if err := c.do(ctx, http.MethodPut, path, req, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// GetCard returns a single saved question (card) by ID including its SQL query.
func (c *Client) GetCard(ctx context.Context, cardID int) (*CardDetail, error) {
	path := fmt.Sprintf("/api/card/%d", cardID)
	var card CardDetail
	if err := c.do(ctx, http.MethodGet, path, nil, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// UpdateDashboardCards updates the layout (position and size) of cards on a dashboard.
func (c *Client) UpdateDashboardCards(ctx context.Context, dashboardID int, cards []map[string]any) error {
	path := fmt.Sprintf("/api/dashboard/%d", dashboardID)
	payload := map[string]any{
		"dashcards": cards,
	}
	return c.do(ctx, http.MethodPut, path, payload, nil)
}

// Search performs a full-text search across questions, dashboards, and collections.
func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error) {
	var resp SearchResponse
	path := "/api/search?q=" + url.QueryEscape(query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
