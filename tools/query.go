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
	DatabaseID int    `json:"database_id" jsonschema:"The ID of the database to query"`
	Query      string `json:"query" jsonschema:"The native SQL query to execute"`
}

// RunQuestionInput is the input for the run_question tool.
type RunQuestionInput struct {
	QuestionID int `json:"question_id" jsonschema:"The ID of the saved question (card) to run"`
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
