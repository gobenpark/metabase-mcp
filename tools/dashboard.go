package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobenpark/metabase-mcp/metabase"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDashboardsInput is the input for list_dashboards.
type ListDashboardsInput struct{}

// GetDashboardInput is the input for get_dashboard.
type GetDashboardInput struct {
	DashboardID int `json:"dashboard_id"`
}

// ListCollectionsInput is the input for list_collections.
type ListCollectionsInput struct{}

// SearchInput is the input for the search tool.
type SearchInput struct {
	Query string `json:"query"`
}

// DeleteDashboardInput is the input for delete_dashboard.
type DeleteDashboardInput struct {
	DashboardID int `json:"dashboard_id"`
}

// ArchiveCollectionInput is the input for archive_collection.
type ArchiveCollectionInput struct {
	CollectionID int `json:"collection_id"`
}

// MoveDashboardInput is the input for move_dashboard.
type MoveDashboardInput struct {
	DashboardID  int `json:"dashboard_id"`
	CollectionID int `json:"collection_id"`
}

// CreateCollectionInput is the input for create_collection.
type CreateCollectionInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ParentID    int    `json:"parent_id,omitempty"`
}

// CreateDashboardInput is the input for create_dashboard.
type CreateDashboardInput struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	CollectionID int    `json:"collection_id,omitempty"`
}

// AddCardToDashboardInput is the input for add_card_to_dashboard.
type AddCardToDashboardInput struct {
	DashboardID int `json:"dashboard_id"`
	CardID      int `json:"card_id"`
	Row         int `json:"row,omitempty"`
	Col         int `json:"col,omitempty"`
	SizeX       int `json:"size_x,omitempty"`
	SizeY       int `json:"size_y,omitempty"`
}

// CardLayoutItem represents the layout of a single card on the dashboard grid.
type CardLayoutItem struct {
	DashcardID int `json:"dashcard_id"`
	Row        int `json:"row"`
	Col        int `json:"col"`
	SizeX      int `json:"size_x"`
	SizeY      int `json:"size_y"`
}

// UpdateDashboardCardsInput is the input for update_dashboard_cards.
type UpdateDashboardCardsInput struct {
	DashboardID int              `json:"dashboard_id"`
	Cards       []CardLayoutItem `json:"cards"`
}

// RegisterDashboardTools registers all dashboard/collection tools on the MCP server.
func RegisterDashboardTools(server *mcp.Server, client *metabase.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dashboards",
		Description: "List all dashboards in this Metabase instance.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDashboardsInput) (*mcp.CallToolResult, any, error) {
		dashboards, err := client.ListDashboards(ctx)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		for _, d := range dashboards {
			desc := d.Description
			if desc == "" {
				desc = "(no description)"
			}
			sb.WriteString(fmt.Sprintf("- [%d] %s — %s\n", d.ID, d.Name, desc))
		}
		if sb.Len() == 0 {
			sb.WriteString("No dashboards found.")
		}
		return textResult(sb.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Get detailed information about a specific dashboard, including its cards/questions.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetDashboardInput) (*mcp.CallToolResult, any, error) {
		dashboard, err := client.GetDashboard(ctx, input.DashboardID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Dashboard: %s (ID: %d)\n", dashboard.Name, dashboard.ID))
		if dashboard.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", dashboard.Description))
		}

		// Calculate grid height from card positions
		maxRow := 0
		for _, dc := range dashboard.Cards {
			if end := dc.Row + dc.SizeY; end > maxRow {
				maxRow = end
			}
		}
		sb.WriteString(fmt.Sprintf("Grid: 24 columns x %d rows (used)\n", maxRow))
		sb.WriteString(fmt.Sprintf("Cards: %d\n\n", len(dashboard.Cards)))

		for _, dc := range dashboard.Cards {
			if dc.CardID == nil {
				continue
			}
			desc := dc.Card.Description
			if desc == "" {
				desc = "(no description)"
			}
			sb.WriteString(fmt.Sprintf("  - [dashcard:%d, card:%d] %s (%s) — %s  [row=%d, col=%d, size=%dx%d]\n",
				dc.ID, dc.Card.ID, dc.Card.Name, dc.Card.Display, desc,
				dc.Row, dc.Col, dc.SizeX, dc.SizeY))
		}
		return textResult(sb.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_collections",
		Description: "List all collections (folders) in this Metabase instance.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListCollectionsInput) (*mcp.CallToolResult, any, error) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		for _, c := range collections {
			if c.Archived {
				continue
			}
			desc := c.Description
			if desc == "" {
				desc = "(no description)"
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s — %s\n", c.ID, c.Name, desc))
		}
		if sb.Len() == 0 {
			sb.WriteString("No collections found.")
		}
		return textResult(sb.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search across questions, dashboards, and collections in Metabase.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
		resp, err := client.Search(ctx, input.Query)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d results:\n\n", resp.Total))
		for _, item := range resp.Data {
			desc := item.Description
			if desc == "" {
				desc = "(no description)"
			}
			sb.WriteString(fmt.Sprintf("- [%s #%d] %s — %s\n", item.Model, item.ID, item.Name, desc))
		}
		if len(resp.Data) == 0 {
			sb.WriteString("No results found.")
		}
		return textResult(sb.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_dashboard",
		Description: "Permanently delete a dashboard from Metabase.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteDashboardInput) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteDashboard(ctx, input.DashboardID); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Dashboard #%d deleted successfully!", input.DashboardID)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "archive_collection",
		Description: "Archive a collection (folder) in Metabase. Archived collections are hidden but can be restored.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ArchiveCollectionInput) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteCollection(ctx, input.CollectionID); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Collection #%d archived successfully!", input.CollectionID)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_dashboard",
		Description: "Move a dashboard to a different collection (folder).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input MoveDashboardInput) (*mcp.CallToolResult, any, error) {
		if err := client.MoveDashboard(ctx, input.DashboardID, input.CollectionID); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Dashboard #%d moved to collection #%d successfully!", input.DashboardID, input.CollectionID)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_collection",
		Description: "Create a new collection (folder) in Metabase. Can be nested under a parent collection.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCollectionInput) (*mcp.CallToolResult, any, error) {
		mbReq := metabase.CreateCollectionRequest{
			Name:        input.Name,
			Description: input.Description,
		}
		if input.ParentID > 0 {
			mbReq.ParentID = &input.ParentID
		}
		collection, err := client.CreateCollection(ctx, mbReq)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Collection created successfully!\n- ID: %s\n- Name: %s", collection.ID, collection.Name)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_dashboard",
		Description: "Create a new dashboard in Metabase. Optionally place it in a collection.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateDashboardInput) (*mcp.CallToolResult, any, error) {
		mbReq := metabase.CreateDashboardRequest{
			Name:        input.Name,
			Description: input.Description,
		}
		if input.CollectionID > 0 {
			mbReq.CollectionID = &input.CollectionID
		}
		dashboard, err := client.CreateDashboard(ctx, mbReq)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Dashboard created successfully!\n- ID: %d\n- Name: %s", dashboard.ID, dashboard.Name)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_dashboard_cards",
		Description: "Update the layout (position and size) of cards on a dashboard. Use get_dashboard first to see current dashcard IDs and layout.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateDashboardCardsInput) (*mcp.CallToolResult, any, error) {
		existing, err := client.GetDashboard(ctx, input.DashboardID)
		if err != nil {
			return errorResult(err), nil, nil
		}

		// Build a map of updates keyed by dashcard ID
		updates := make(map[int]CardLayoutItem)
		for _, c := range input.Cards {
			updates[c.DashcardID] = c
		}

		// Merge updates with existing cards
		cards := make([]map[string]any, 0, len(existing.Cards))
		for _, dc := range existing.Cards {
			entry := map[string]any{
				"id":      dc.ID,
				"card_id": dc.CardID,
				"row":     dc.Row,
				"col":     dc.Col,
				"size_x":  dc.SizeX,
				"size_y":  dc.SizeY,
			}
			if u, ok := updates[dc.ID]; ok {
				entry["row"] = u.Row
				entry["col"] = u.Col
				entry["size_x"] = u.SizeX
				entry["size_y"] = u.SizeY
			}
			cards = append(cards, entry)
		}

		if err := client.UpdateDashboardCards(ctx, input.DashboardID, cards); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Updated layout of %d card(s) on dashboard #%d!", len(input.Cards), input.DashboardID)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_card_to_dashboard",
		Description: "Add an existing saved question (card) to a dashboard.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddCardToDashboardInput) (*mcp.CallToolResult, any, error) {
		sizeX := input.SizeX
		if sizeX == 0 {
			sizeX = 6
		}
		sizeY := input.SizeY
		if sizeY == 0 {
			sizeY = 4
		}
		mbReq := metabase.AddCardToDashboardRequest{
			CardID: input.CardID,
			Row:    input.Row,
			Col:    input.Col,
			SizeX:  sizeX,
			SizeY:  sizeY,
		}
		if err := client.AddCardToDashboard(ctx, input.DashboardID, mbReq); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Card #%d added to dashboard #%d successfully!", input.CardID, input.DashboardID)), nil, nil
	})
}
