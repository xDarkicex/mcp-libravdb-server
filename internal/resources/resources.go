package resources

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers static resources and an empty template list (required for Cline compatibility).
func RegisterAll(server *mcp.Server) {
	handler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		switch uri {
		case "memory://collections":
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      uri,
					MIMEType: "application/json",
					Text:     `{"collections":[],"note":"collections are populated from the daemon at request time"}`,
				}},
			}, nil
		case "memory://kinds":
			kinds := []map[string]string{
				{"kind": "identity", "description": "Self-referential declarations about who someone is"},
				{"kind": "constraint", "description": "Rules, limits, or requirements that must be followed"},
				{"kind": "decision", "description": "Choices made with rationale and context"},
				{"kind": "fact", "description": "Descriptive statements of what is true"},
				{"kind": "preference", "description": "Affinities, aversions, or stylistic choices"},
				{"kind": "episode", "description": "Sequences of events with temporal ordering"},
			}
			b, _ := json.Marshal(kinds)
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(b),
				}},
			}, nil
		case "memory://status":
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      uri,
					MIMEType: "application/json",
					Text:     `{"status":"see memory.stats tool for live daemon status"}`,
				}},
			}, nil
		default:
			return nil, mcp.ResourceNotFoundError(uri)
		}
	}

	server.AddResource(&mcp.Resource{
		URI:         "memory://collections",
		Name:        "Memory Collections",
		Description: "All collection identifiers with per-collection record counts",
		MIMEType:    "application/json",
	}, handler)
	server.AddResource(&mcp.Resource{
		URI:         "memory://kinds",
		Name:        "Cognitive Kinds",
		Description: "Cognitive kind registry (identity, constraint, decision, fact, preference, episode)",
		MIMEType:    "application/json",
	}, handler)
	server.AddResource(&mcp.Resource{
		URI:         "memory://status",
		Name:        "Daemon Status",
		Description: "Live daemon health, embedding profile, memory counts, gating threshold",
		MIMEType:    "application/json",
	}, handler)

	// Empty template list — required for Cline compatibility
	// Cline unconditionally calls resources/templates/list regardless of capabilities
	server.AddResourceTemplate(&mcp.ResourceTemplate{}, nil)
}
