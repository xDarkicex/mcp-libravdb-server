package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func TestRecall_WithGating(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	fake.SearchResults = []*ipcv1.SearchResult{
		{Id: "id-1", Score: 0.9, Text: "result one", Version: 1},
	}

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.recall",
		Arguments: map[string]any{"query": "test query"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestRecall_KindFilter(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	fake.SearchResults = []*ipcv1.SearchResult{
		{Id: "id-1", Score: 0.9, Text: "decision record", MetadataJson: []byte(`{"memory_kind":"decision"}`)},
		{Id: "id-2", Score: 0.8, Text: "fact record", MetadataJson: []byte(`{"memory_kind":"fact"}`)},
	}

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.recall",
		Arguments: map[string]any{"query": "test", "kind": "decision"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestRecall_BackendError(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.Error = assert.AnError

	session := startServer(t, client)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.recall", Arguments: map[string]any{"query": "test"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestKindMatches_InvalidJSON(t *testing.T) {
	if kindMatches([]byte("not json"), "decision") {
		t.Fatal("invalid JSON should not match")
	}
}

func TestKindMatches_EmptyKind(t *testing.T) {
	if !kindMatches(nil, "") {
		t.Fatal("empty kind should match everything")
	}
}

func TestKindMatches_WrongKind(t *testing.T) {
	if kindMatches([]byte(`{"memory_kind":"fact"}`), "decision") {
		t.Fatal("wrong kind should not match")
	}
}

func TestRecall_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.recall",
		Arguments: map[string]any{"query": "test"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
