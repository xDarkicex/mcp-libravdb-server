package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type GraphArgs struct {
	RecordID string `json:"record_id" jsonschema:"memory ID to walk causal edges from. The daemon traverses why_ids, how_ids, and hop_targets from this record."`
	MaxDepth int32  `json:"max_depth,omitempty" jsonschema:"maximum graph walk depth (default 3)"`
}

func registerGraph(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.graph",
		Description: "Walk the causal graph from a memory record. Returns connected records with why (causal parent), how (procedural child), and hop (conceptual) edges. The daemon's TopoRegistry does the traversal.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in GraphArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}
		if in.MaxDepth == 0 {
			in.MaxDepth = 3
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		resp, err := deps.Client.ExpandSummary(ctx, &ipcv1.ExpandSummaryRequest{
			RecordId: in.RecordID,
			MaxDepth: in.MaxDepth,
		})
		if err != nil {
			deps.Logger.Error("memory.graph failed", "record_id", in.RecordID, "err", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "INVALID_MEMORY_ID: " + err.Error()},
				},
			}, nil, nil
		}

		connected := make([]ConnectedRecord, len(resp.Connected))
		for i, c := range resp.Connected {
			connected[i] = ConnectedRecord{
				RecordID: c.RecordId, Text: c.Text,
				Depth: c.Depth, EdgeWeight: c.EdgeWeight, EdgeType: c.EdgeType,
			}
		}

		return &mcp.CallToolResult{
			StructuredContent: GraphResult{
				RecordID:  in.RecordID,
				WhyIDs:    resp.WhyIds,
				HowIDs:    resp.HowIds,
				HopTarget: resp.HopTargets,
				Connected: connected,
			},
		}, nil, nil
	})
}
