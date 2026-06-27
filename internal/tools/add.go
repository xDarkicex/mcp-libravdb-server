package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type AddArgs struct {
	Collection string `json:"collection" jsonschema:"collection to add the memory to"`
	Text       string `json:"text" jsonschema:"memory content text"`
	ID         string `json:"id,omitempty" jsonschema:"optional memory ID (auto-generated UUID if omitted)"`
}

func registerAdd(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.add",
		Description: "Add a memory record to a collection. The daemon handles cognitive classification, deontic analysis, and embedding automatically.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in AddArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}

		id := in.ID
		if id == "" {
			id = uuid.New().String()
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		resp, err := deps.Client.InsertText(ctx, &ipcv1.InsertTextRequest{
			Collection: in.Collection,
			Id:         id,
			Text:       in.Text,
		})
		if err != nil {
			deps.Logger.Error("memory.add failed", "collection", in.Collection, "err", err)
			return backendUnavailable(), nil, nil
		}

		return &mcp.CallToolResult{
			StructuredContent: AddResult{ID: id, Collection: in.Collection, OK: resp.Ok},
		}, nil, nil
	})
}
