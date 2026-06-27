package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete_OK(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.DeleteOK = true

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.delete",
		Arguments: map[string]any{
			"collection": "test",
			"id":         "id-to-delete",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "id-to-delete", fake.LastDelete.Id)
	assert.Equal(t, "test", fake.LastDelete.Collection)
}

func TestDelete_BackendError(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.Error = assert.AnError

	session := startServer(t, client)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.delete", Arguments: map[string]any{"collection": "test", "id": "any"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestDelete_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.delete",
		Arguments: map[string]any{"collection": "test", "id": "any"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
