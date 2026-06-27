package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type PredictArgs struct {
	Query      string `json:"query" jsonschema:"text describing the current situation or decision"`
	Collection string `json:"collection,omitempty" jsonschema:"scope to a specific collection"`
	MaxDepth   int32  `json:"max_depth,omitempty" jsonschema:"graph walk depth from matched records (default 3)"`
	Limit      int32  `json:"limit,omitempty" jsonschema:"maximum predictions returned (default 10)"`
}

func registerPredict(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.predict",
		Description: "Find memories likely relevant next. Searches for state nodes matching the query, then delegates graph traversal to the daemon's ExpandSummary for each match.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in PredictArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}
		if in.MaxDepth == 0 {
			in.MaxDepth = 3
		}
		if in.Limit == 0 {
			in.Limit = 10
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		predictions, err := collectPredictions(ctx, deps.Client, in)
		if err != nil {
			return backendUnavailable(), nil, nil
		}
		if len(predictions) == 0 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "NO_PREDICTIONS_FOUND: no connected records found from matched state nodes"},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			StructuredContent: PredictResult{Query: in.Query, Predictions: predictions},
		}, nil, nil
	})
}

func collectPredictions(ctx context.Context, client ipcv1.LibravDBClient, in PredictArgs) ([]Prediction, error) {
	searchResp, err := client.SearchText(ctx, &ipcv1.SearchTextRequest{
		Collection: in.Collection,
		Text:       in.Query,
		K:          5,
	})
	if err != nil {
		return nil, err
	}

	out := make([]Prediction, 0, in.Limit)
	for _, r := range searchResp.Results {
		if int32(len(out)) >= in.Limit {
			break
		}
		walkPredictions(ctx, client, r.Id, in.MaxDepth, in.Limit, &out)
	}
	return out, nil
}

func walkPredictions(ctx context.Context, client ipcv1.LibravDBClient, recordID string, maxDepth, limit int32, out *[]Prediction) {
	resp, err := client.ExpandSummary(ctx, &ipcv1.ExpandSummaryRequest{
		RecordId: recordID,
		MaxDepth: maxDepth,
	})
	if err != nil {
		return
	}
	for _, c := range resp.Connected {
		if int32(len(*out)) >= limit {
			return
		}
		score := 1.0 / (1.0 + float64(c.Depth))
		if score >= 0.25 {
			*out = append(*out, Prediction{
				ID: c.RecordId, Text: c.Text, SourceNodeID: recordID,
				Depth: c.Depth, CausalScore: score, EdgeType: c.EdgeType,
			})
		}
	}
}
