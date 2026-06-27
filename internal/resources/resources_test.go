package resources

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterAll(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, &mcp.ServerOptions{
		HasResources: true,
	})
	RegisterAll(server)

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() {
		_, _ = server.Connect(t.Context(), serverTrans, nil)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	resources, err := session.ListResources(t.Context(), nil)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(resources.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources.Resources))
	}

	uris := map[string]bool{}
	for _, r := range resources.Resources {
		uris[r.URI] = true
	}
	for _, uri := range []string{"memory://collections", "memory://kinds", "memory://status"} {
		if !uris[uri] {
			t.Errorf("missing resource: %s", uri)
		}
	}
}

func TestResourcesRead(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, &mcp.ServerOptions{
		HasResources: true,
	})
	RegisterAll(server)

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() {
		_, _ = server.Connect(t.Context(), serverTrans, nil)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	for _, uri := range []string{"memory://collections", "memory://kinds", "memory://status"} {
		result, err := session.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: uri})
		if err != nil {
			t.Errorf("read %s: %v", uri, err)
			continue
		}
		if len(result.Contents) == 0 {
			t.Errorf("expected content for %s", uri)
		}
	}
}

func TestResourceTemplatesList(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, &mcp.ServerOptions{
		HasResources: true,
	})
	RegisterAll(server)

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() { _, _ = server.Connect(t.Context(), serverTrans, nil) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	templates, err := session.ListResourceTemplates(t.Context(), nil)
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	// Should return empty list for Cline compatibility
	if templates == nil {
		t.Fatal("expected non-nil template list")
	}
}

func TestResourcesRead_NotFound(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, &mcp.ServerOptions{
		HasResources: true,
	})
	RegisterAll(server)

	clientTrans, serverTrans := mcp.NewInMemoryTransports()
	go func() {
		_, _ = server.Connect(t.Context(), serverTrans, nil)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTrans, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	_, err = session.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: "memory://nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown resource")
	}
}
