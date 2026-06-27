package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type SearchArgs struct {
	Query       string   `json:"query" jsonschema:"semantic search query text"`
	Collection  string   `json:"collection,omitempty" jsonschema:"scope search to a specific collection"`
	Collections []string `json:"collections,omitempty" jsonschema:"search across multiple collections (alternative to collection)"`
	Limit       int32    `json:"limit,omitempty" jsonschema:"maximum results (default 20)"`
}

func registerSearch(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.search",
		Description: "Search memory using semantic similarity. Finds records matching the query text.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in SearchArgs) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}

		k := in.Limit
		if k == 0 {
			k = 20
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		resp, err := search(ctx, deps.Client, in.Query, in.Collection, in.Collections, k)
		if err != nil {
			deps.Logger.Error("memory.search failed", "err", err)
			return backendUnavailable(), nil, nil
		}

		results := make([]SearchResult, len(resp.Results))
		for i, r := range resp.Results {
			results[i] = SearchResult{
				ID:      r.Id,
				Score:   r.Score,
				Text:    r.Text,
				Meta:    string(r.MetadataJson),
				Version: r.Version,
			}
		}

		return &mcp.CallToolResult{
			StructuredContent: results,
		}, nil, nil
	})
}

func search(ctx context.Context, client ipcv1.LibravDBClient, query, collection string, collections []string, k int32) (*ipcv1.SearchTextResponse, error) {
	if collection != "" {
		return client.SearchText(ctx, &ipcv1.SearchTextRequest{
			Collection: collection, Text: query, K: k,
		})
	}
	if len(collections) > 0 {
		return client.SearchTextCollections(ctx, &ipcv1.SearchTextCollectionsRequest{
			Collections: collections, Text: query, K: k,
		})
	}
	return client.SearchText(ctx, &ipcv1.SearchTextRequest{Text: query, K: k})
}
