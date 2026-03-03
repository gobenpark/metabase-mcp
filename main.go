package main

import (
	"context"
	"log"
	"os"

	"github.com/gobenpark/metabase-mcp/metabase"
	"github.com/gobenpark/metabase-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	log.SetOutput(os.Stderr)

	client, err := metabase.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize Metabase client: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "metabase-mcp",
		Version: version,
	}, nil)

	tools.RegisterQueryTools(server, client)
	tools.RegisterDashboardTools(server, client)

	log.Println("Starting metabase-mcp server (stdio)...")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
