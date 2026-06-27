package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStats_OK(t *testing.T) {
	_, client, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.stats",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestStats_BackendError(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.Error = assert.AnError

	session := startServer(t, client)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "memory.stats"})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestStats_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.stats",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
