package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func TestSearch_SingleCollection(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	fake.SearchResults = []*ipcv1.SearchResult{
		{Id: "id-1", Score: 0.95, Text: "result one", Version: 1},
	}

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.search",
		Arguments: map[string]any{
			"query":      "test query",
			"collection": "test-collection",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "test-collection", fake.LastSearch.Collection)
	assert.Equal(t, "test query", fake.LastSearch.Text)
}

func TestSearch_MultiCollection(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	fake.SearchResults = []*ipcv1.SearchResult{
		{Id: "id-1", Score: 0.9, Text: "result"},
	}

	session := startServer(t, client)

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.search",
		Arguments: map[string]any{"query": "test", "collections": []any{"a", "b"}},
	})
	require.NoError(t, err)
}

func TestSearch_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.search",
		Arguments: map[string]any{"query": "test"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSearch_DefaultLimit(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, client)

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.search",
		Arguments: map[string]any{"query": "test"},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(20), fake.LastSearch.K)
}
