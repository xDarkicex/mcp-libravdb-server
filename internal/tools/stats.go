package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func registerStats(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory.stats",
		Description: "Get memory system statistics: total nodes, counts by cognitive kind and tier, daemon status, and embedding profile.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ any,
	) (*mcp.CallToolResult, any, error) {
		if !deps.BackendHealthy {
			return backendUnavailable(), nil, nil
		}

		ctx, cancel := context.WithTimeout(ctx, deps.BackendTimeout)
		defer cancel()

		metrics, err := deps.Client.CognitiveMetrics(ctx, &ipcv1.CognitiveMetricsRequest{})
		if err != nil {
			deps.Logger.Error("memory.stats metrics failed", "err", err)
			return backendUnavailable(), nil, nil
		}

		result := StatsResult{
			TotalNodes: metrics.TotalNodes,
			CountsByKind: KindCounts{
				Identity: metrics.Identity, Constraint: metrics.Constraint,
				Decision: metrics.Decision, Fact: metrics.Fact,
				Preference: metrics.Preference, Episode: metrics.Episode,
			},
			CountsByTier: TierCounts{
				Hard: metrics.TierHard, Soft: metrics.TierSoft, Variant: metrics.TierVariant,
			},
			CountsByHeading: HeadingCounts{
				Identity: metrics.HeadingIdentity, Constraint: metrics.HeadingConstraint,
				Workflow: metrics.HeadingWorkflow, Background: metrics.HeadingBackground,
				Preferences: metrics.HeadingPreferences,
			},
		}

		status, err := deps.Client.Status(ctx, &ipcv1.MemoryStatusRequest{})
		if err == nil {
			result.DaemonStatus = &DaemonStatusInfo{
				OK: status.Ok, TurnCount: status.TurnCount,
				MemoryCount: status.MemoryCount, GatingThreshold: status.GatingThreshold,
				EmbeddingProfile: status.EmbeddingProfile,
			}
		} else {
			deps.Logger.Warn("memory.stats status failed, using metrics only", "err", err)
		}

		return &mcp.CallToolResult{
			StructuredContent: result,
		}, nil, nil
	})
}
