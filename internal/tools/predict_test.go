package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func TestPredict_Found(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	fake.SearchResults = []*ipcv1.SearchResult{
		{Id: "state-1", Score: 0.9, Text: "state node"},
	}

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.predict",
		Arguments: map[string]any{"query": "current situation"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError, "should find predictions from connected records")
}

func TestPredict_NoResults(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()

	// No connected records in response → no predictions
	fake.GraphResponse = &ipcv1.ExpandSummaryResponse{}

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.predict",
		Arguments: map[string]any{"query": "nothing matches"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestPredict_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.predict",
		Arguments: map[string]any{"query": "test"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
