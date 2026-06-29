package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type RecallArgs struct {
	Query      string `json:"query" jsonschema:"semantic search query"`
	Collection string `json:"collection,omitempty" jsonschema:"scope to a specific collection"`
	Kind       string `json:"kind,omitempty" jsonschema:"optional cognitive kind filter (identity, constraint, decision, fact, preference, episode)"`
	Limit      int32  `json:"limit,omitempty" jsonschema:"maximum results (default 20)"`
}

func registerRecall(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.recall",
		Description: "Search memory with gating enrichment and optional cognitive kind filtering. Returns search results with deontic gating scores.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in RecallArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}
		if in.Limit == 0 {
			in.Limit = 20
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		results, err := searchAndEnrich(ctx, deps.Client, in)
		if err != nil {
			deps.Logger.Error("memory.recall failed", "err", err)
			return backendUnavailable(), nil, nil
		}

		return &mcp.CallToolResult{
			StructuredContent: RecallResponse{Results: results},
		}, nil, nil
	})
}

func searchAndEnrich(ctx context.Context, client ipcv1.LibravDBClient, in RecallArgs) ([]RecallResult, error) {
	searchResp, err := search(ctx, client, in.Query, in.Collection, nil, in.Limit*2)
	if err != nil {
		return nil, err
	}

	results := make([]RecallResult, 0, in.Limit)
	for _, r := range searchResp.Results {
		if int32(len(results)) >= in.Limit {
			break
		}
		if !kindMatches(r.MetadataJson, in.Kind) {
			continue
		}

		results = append(results, RecallResult{
			ID:      r.Id,
			Score:   r.Score,
			Text:    r.Text,
			Meta:    string(r.MetadataJson),
			Version: r.Version,
			Gating:  fetchGating(ctx, client, r.Text),
		})
	}
	return results, nil
}

func kindMatches(metadataJSON []byte, kind string) bool {
	if kind == "" {
		return true
	}
	var meta map[string]interface{}
	if json.Unmarshal(metadataJSON, &meta) != nil {
		return false
	}
	k, _ := meta["memory_kind"].(string)
	return k == kind
}

func fetchGating(ctx context.Context, client ipcv1.LibravDBClient, text string) *GatingScores {
	resp, err := client.GatingScalar(ctx, &ipcv1.GatingScalarRequest{Text: text})
	if err != nil {
		return nil
	}
	return &GatingScores{
		G: resp.G, H: resp.H, R: resp.R, DNL: resp.D,
		P: resp.P, A: resp.A, DTech: resp.Dtech,
		GConv: resp.Gconv, GTech: resp.Gtech,
	}
}
