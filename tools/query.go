package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobenpark/metabase-mcp/metabase"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExecuteQueryInput is the input for the execute_query tool.
type ExecuteQueryInput struct {
	DatabaseID int    `json:"database_id"`
	Query      string `json:"query"`
}

// RunQuestionInput is the input for the run_question tool.
type RunQuestionInput struct {
	QuestionID int `json:"question_id"`
}

// CreateCardInput is the input for create_card.
type CreateCardInput struct {
	Name                  string         `json:"name"`
	Description           string         `json:"description,omitempty"`
	DatabaseID            int            `json:"database_id"`
	Query                 string         `json:"query"`
	Display               string         `json:"display,omitempty"`
	CollectionID          int            `json:"collection_id,omitempty"`
	VisualizationSettings map[string]any `json:"visualization_settings,omitempty"`
}

// UpdateCardDisplayInput is the input for update_card_display.
type UpdateCardDisplayInput struct {
	CardID  int    `json:"card_id"`
	Display string `json:"display"`
}

// UpdateCardInput is the input for update_card.
type UpdateCardInput struct {
	CardID                int            `json:"card_id"`
	Query                 string         `json:"query,omitempty"`
	DatabaseID            int            `json:"database_id,omitempty"`
	Name                  string         `json:"name,omitempty"`
	Description           string         `json:"description,omitempty"`
	VisualizationSettings map[string]any `json:"visualization_settings,omitempty"`
}

// GetCardInput is the input for get_card.
type GetCardInput struct {
	CardID int `json:"card_id"`
}

// ListDatabasesInput is the input for list_databases (no params needed).
type ListDatabasesInput struct{}

// RegisterQueryTools registers all query-related tools on the MCP server.
func RegisterQueryTools(server *mcp.Server, client *metabase.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_query",
		Description: "Execute a native SQL query against a Metabase database. Returns column names and rows.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExecuteQueryInput) (*mcp.CallToolResult, any, error) {
		result, err := client.ExecuteNativeQuery(ctx, input.DatabaseID, input.Query)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(formatQueryResult(result)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_question",
		Description: "Run a saved Metabase question (card) by its ID and return the results.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RunQuestionInput) (*mcp.CallToolResult, any, error) {
		result, err := client.RunQuestion(ctx, input.QuestionID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(formatQueryResult(result)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_card",
		Description: "Create a new saved question (card) with a native SQL query. You can specify the chart type (display) when creating.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCardInput) (*mcp.CallToolResult, any, error) {
		display := input.Display
		if display == "" {
			display = "table"
		}
		vizSettings := input.VisualizationSettings
		if vizSettings == nil {
			vizSettings = map[string]any{}
		}
		mbReq := metabase.CreateCardRequest{
			Name:        input.Name,
			Description: input.Description,
			Display:     display,
			DatasetQuery: metabase.DatasetQuery{
				Database: input.DatabaseID,
				Type:     "native",
				Native:   &metabase.NativeQuery{Query: input.Query},
			},
			VisualizationSettings: vizSettings,
		}
		if input.CollectionID > 0 {
			mbReq.CollectionID = &input.CollectionID
		}
		card, err := client.CreateCard(ctx, mbReq)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Card created successfully!\n- ID: %d\n- Name: %s\n- Display: %s", card.ID, card.Name, display)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_card_display",
		Description: "Change the chart/visualization type of a saved question (card). Supported types: table, bar, line, pie, scalar, row, area, combo, pivot, funnel, map, scatter, waterfall, progress, gauge.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateCardDisplayInput) (*mcp.CallToolResult, any, error) {
		if err := client.UpdateCardDisplay(ctx, input.CardID, input.Display); err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Card #%d display changed to '%s'!", input.CardID, input.Display)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_card",
		Description: "Get a saved question (card) by ID, including its current SQL query and metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetCardInput) (*mcp.CallToolResult, any, error) {
		card, err := client.GetCard(ctx, input.CardID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Card #%d: %s\n", card.ID, card.Name))
		if card.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", card.Description))
		}
		sb.WriteString(fmt.Sprintf("Display: %s\n", card.Display))
		// Extract SQL from either the legacy native format or new stages format
		sql := ""
		if card.DatasetQuery.Native != nil {
			sql = card.DatasetQuery.Native.Query
		} else if len(card.DatasetQuery.Stages) > 0 {
			sql = card.DatasetQuery.Stages[0].Native
		}
		if sql != "" {
			sb.WriteString(fmt.Sprintf("\nSQL:\n%s", sql))
		}
		return textResult(sb.String()), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_card",
		Description: "Update a saved question (card): change its SQL query, name, or description. Provide only the fields you want to change.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateCardInput) (*mcp.CallToolResult, any, error) {
		if input.Query != "" && input.DatabaseID == 0 {
			return errorResult(fmt.Errorf("database_id is required when updating query")), nil, nil
		}
		mbReq := metabase.UpdateCardRequest{
			Name:                  input.Name,
			Description:           input.Description,
			VisualizationSettings: input.VisualizationSettings,
		}
		if input.Query != "" {
			mbReq.DatasetQuery = &metabase.DatasetQuery{
				Database: input.DatabaseID,
				Type:     "native",
				Native:   &metabase.NativeQuery{Query: input.Query},
			}
		}
		card, err := client.UpdateCard(ctx, input.CardID, mbReq)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Card #%d updated successfully!\n- Name: %s", card.ID, card.Name)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_databases",
		Description: "List all databases connected to this Metabase instance.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDatabasesInput) (*mcp.CallToolResult, any, error) {
		databases, err := client.ListDatabases(ctx)
		if err != nil {
			return errorResult(err), nil, nil
		}
		var sb strings.Builder
		for _, db := range databases {
			sb.WriteString(fmt.Sprintf("- [%d] %s (engine: %s)\n", db.ID, db.Name, db.Engine))
		}
		if sb.Len() == 0 {
			sb.WriteString("No databases found.")
		}
		return textResult(sb.String()), nil, nil
	})
}

func formatQueryResult(result *metabase.QueryResult) string {
	var sb strings.Builder

	// Column headers
	colNames := make([]string, len(result.Data.Cols))
	for i, col := range result.Data.Cols {
		colNames[i] = col.DisplayName
	}
	sb.WriteString(strings.Join(colNames, " | "))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(sb.String())-1))
	sb.WriteString("\n")

	// Rows (limit to 100 for readability)
	maxRows := 100
	for i, row := range result.Data.Rows {
		if i >= maxRows {
			sb.WriteString(fmt.Sprintf("\n... (%d more rows truncated)", len(result.Data.Rows)-maxRows))
			break
		}
		vals := make([]string, len(row))
		for j, v := range row {
			if v == nil {
				vals[j] = "NULL"
			} else {
				vals[j] = fmt.Sprintf("%v", v)
			}
		}
		sb.WriteString(strings.Join(vals, " | "))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\nTotal rows: %d", len(result.Data.Rows)))
	return sb.String()
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Error: %s", err.Error())},
		},
		IsError: true,
	}
}
