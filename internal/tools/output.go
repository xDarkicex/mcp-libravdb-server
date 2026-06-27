package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func backendUnavailable() *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: "MEMORY_BACKEND_UNAVAILABLE: daemon is unreachable"},
		},
	}
}

// All output types are passed directly to CallToolResult.StructuredContent.
// The MCP SDK handles serialization — we never call json.Marshal on the hot path.

type SearchResult struct {
	ID      string  `json:"id"`
	Score   float64 `json:"score"`
	Text    string  `json:"text"`
	Meta    string  `json:"metadata_json"`
	Version uint64  `json:"version"`
}

type AddResult struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	OK         bool   `json:"ok"`
}

type DeleteResult struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	OK         bool   `json:"ok"`
}

type GatingScores struct {
	G     float64 `json:"g"`
	H     float64 `json:"h"`
	R     float64 `json:"r"`
	DNL   float64 `json:"dnl"`
	P     float64 `json:"p"`
	A     float64 `json:"a"`
	DTech float64 `json:"dtech"`
	GConv float64 `json:"gconv"`
	GTech float64 `json:"gtech"`
}

type RecallResult struct {
	ID      string        `json:"id"`
	Score   float64       `json:"score"`
	Text    string        `json:"text"`
	Meta    string        `json:"metadata_json"`
	Version uint64        `json:"version"`
	Gating  *GatingScores `json:"gating,omitempty"`
}

type ConnectedRecord struct {
	RecordID   string  `json:"record_id"`
	Text       string  `json:"text"`
	Depth      int32   `json:"depth"`
	EdgeWeight float64 `json:"edge_weight"`
	EdgeType   string  `json:"edge_type"`
}

type GraphResult struct {
	RecordID  string            `json:"record_id"`
	WhyIDs    []string          `json:"why_ids"`
	HowIDs    []string          `json:"how_ids"`
	HopTarget []string          `json:"hop_targets"`
	Connected []ConnectedRecord `json:"connected"`
}

type Prediction struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	SourceNodeID string  `json:"source_node_id"`
	Depth        int32   `json:"depth"`
	CausalScore  float64 `json:"causal_score"`
	EdgeType     string  `json:"edge_type"`
}

type PredictResult struct {
	Query       string       `json:"query"`
	Predictions []Prediction `json:"predictions"`
}

type StatsResult struct {
	TotalNodes     int64            `json:"total_nodes"`
	CountsByKind   KindCounts       `json:"counts_by_kind"`
	CountsByTier   TierCounts       `json:"counts_by_tier"`
	CountsByHeading HeadingCounts    `json:"counts_by_heading"`
	DaemonStatus   *DaemonStatusInfo `json:"daemon_status,omitempty"`
}

type KindCounts struct {
	Identity   int64 `json:"identity"`
	Constraint int64 `json:"constraint"`
	Decision   int64 `json:"decision"`
	Fact       int64 `json:"fact"`
	Preference int64 `json:"preference"`
	Episode    int64 `json:"episode"`
}

type TierCounts struct {
	Hard    int64 `json:"hard"`
	Soft    int64 `json:"soft"`
	Variant int64 `json:"variant"`
}

type HeadingCounts struct {
	Identity    int64 `json:"identity"`
	Constraint  int64 `json:"constraint"`
	Workflow    int64 `json:"workflow"`
	Background  int64 `json:"background"`
	Preferences int64 `json:"preferences"`
}

type DaemonStatusInfo struct {
	OK               bool    `json:"ok"`
	TurnCount        int32   `json:"turn_count"`
	MemoryCount      int32   `json:"memory_count"`
	GatingThreshold  float64 `json:"gating_threshold"`
	EmbeddingProfile string  `json:"embedding_profile"`
}
