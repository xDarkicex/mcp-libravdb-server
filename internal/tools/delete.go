package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type DeleteArgs struct {
	Collection string `json:"collection" jsonschema:"collection containing the memory"`
	ID         string `json:"id" jsonschema:"memory ID to delete"`
}

func registerDelete(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.delete",
		Description: "Delete a single memory record by ID. Single-ID deletion only.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in DeleteArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		resp, err := deps.Client.Delete(ctx, &ipcv1.DeleteRequest{
			Collection: in.Collection,
			Id:         in.ID,
		})
		if err != nil {
			deps.Logger.Error("memory.delete failed", "collection", in.Collection, "id", in.ID, "err", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "INVALID_MEMORY_ID: " + err.Error()},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			StructuredContent: DeleteResult{ID: in.ID, Collection: in.Collection, OK: resp.Ok},
		}, nil, nil
	})
}
