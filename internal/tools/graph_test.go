package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraph_Walk(t *testing.T) {
	_, client, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.graph",
		Arguments: map[string]any{"record_id": "record-1", "max_depth": 3},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestGraph_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.graph",
		Arguments: map[string]any{"record_id": "any"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestGraph_DefaultDepth(t *testing.T) {
	_, client, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, client)

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.graph",
		Arguments: map[string]any{"record_id": "record-1"},
	})
	require.NoError(t, err)
}
