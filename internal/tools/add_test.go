package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_Insert(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.InsertOK = true

	session := startServer(t, client)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.add",
		Arguments: map[string]any{
			"collection": "test-collection",
			"text":       "memory content",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "test-collection", fake.LastInsert.Collection)
	assert.Equal(t, "memory content", fake.LastInsert.Text)
	assert.NotEmpty(t, fake.LastInsert.Id, "should auto-generate UUID")
}

func TestAdd_ExplicitID(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.InsertOK = true

	session := startServer(t, client)

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.add",
		Arguments: map[string]any{
			"collection": "test",
			"text":       "content",
			"id":         "my-custom-id",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "my-custom-id", fake.LastInsert.Id)
}

func TestAdd_BackendError(t *testing.T) {
	fake, client, cleanup := setupTest(t)
	defer cleanup()
	fake.Error = assert.AnError

	session := startServer(t, client)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "memory.add", Arguments: map[string]any{"collection": "test", "text": "content"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestAdd_Degraded(t *testing.T) {
	_, _, cleanup := setupTest(t)
	defer cleanup()

	session := startServer(t, nil)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "memory.add",
		Arguments: map[string]any{"collection": "test", "text": "content"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
