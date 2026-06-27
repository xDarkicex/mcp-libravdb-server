package testutil

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

// FakeBackend is an in-process fake libravdb daemon for tests.
type FakeBackend struct {
	ipcv1.UnimplementedLibravDBServer

	SearchResults   []*ipcv1.SearchResult
	InsertOK        bool
	DeleteOK        bool
	GatingResponse  *ipcv1.GatingScalarResponse
	HealthOK        bool
	MetricsResponse *ipcv1.CognitiveMetricsResponse
	StatusResponse  *ipcv1.MemoryStatusResponse
	GraphResponse   *ipcv1.ExpandSummaryResponse

	LastSearch   *ipcv1.SearchTextRequest
	LastInsert   *ipcv1.InsertTextRequest
	LastDelete   *ipcv1.DeleteRequest
}

func (f *FakeBackend) Health(ctx context.Context, req *ipcv1.HealthRequest) (*ipcv1.HealthResponse, error) {
	return &ipcv1.HealthResponse{Ok: f.HealthOK}, nil
}

func (f *FakeBackend) SearchText(ctx context.Context, req *ipcv1.SearchTextRequest) (*ipcv1.SearchTextResponse, error) {
	f.LastSearch = req
	return &ipcv1.SearchTextResponse{Results: f.SearchResults}, nil
}

func (f *FakeBackend) SearchTextCollections(ctx context.Context, req *ipcv1.SearchTextCollectionsRequest) (*ipcv1.SearchTextResponse, error) {
	return &ipcv1.SearchTextResponse{Results: f.SearchResults}, nil
}

func (f *FakeBackend) InsertText(ctx context.Context, req *ipcv1.InsertTextRequest) (*ipcv1.InsertTextResponse, error) {
	f.LastInsert = req
	return &ipcv1.InsertTextResponse{Ok: f.InsertOK}, nil
}

func (f *FakeBackend) Delete(ctx context.Context, req *ipcv1.DeleteRequest) (*ipcv1.DeleteResponse, error) {
	f.LastDelete = req
	return &ipcv1.DeleteResponse{Ok: f.DeleteOK}, nil
}

func (f *FakeBackend) GatingScalar(ctx context.Context, req *ipcv1.GatingScalarRequest) (*ipcv1.GatingScalarResponse, error) {
	if f.GatingResponse != nil {
		return f.GatingResponse, nil
	}
	return &ipcv1.GatingScalarResponse{G: 0.5, Gconv: 0.4, Gtech: 0.6}, nil
}

func (f *FakeBackend) CognitiveMetrics(ctx context.Context, req *ipcv1.CognitiveMetricsRequest) (*ipcv1.CognitiveMetricsResponse, error) {
	if f.MetricsResponse != nil {
		return f.MetricsResponse, nil
	}
	return &ipcv1.CognitiveMetricsResponse{
		TotalNodes: 100, Identity: 10, Constraint: 15, Decision: 25,
		Fact: 30, Preference: 12, Episode: 8,
		TierHard: 40, TierSoft: 35, TierVariant: 25,
		HeadingIdentity: 10, HeadingConstraint: 15, HeadingWorkflow: 30,
		HeadingBackground: 25, HeadingPreferences: 20,
	}, nil
}

func (f *FakeBackend) Status(ctx context.Context, req *ipcv1.MemoryStatusRequest) (*ipcv1.MemoryStatusResponse, error) {
	if f.StatusResponse != nil {
		return f.StatusResponse, nil
	}
	return &ipcv1.MemoryStatusResponse{
		Ok: true, TurnCount: 42, MemoryCount: 100, GatingThreshold: 0.35,
		EmbeddingProfile: "test-embedder",
	}, nil
}

func (f *FakeBackend) ExpandSummary(ctx context.Context, req *ipcv1.ExpandSummaryRequest) (*ipcv1.ExpandSummaryResponse, error) {
	if f.GraphResponse != nil {
		return f.GraphResponse, nil
	}
	return &ipcv1.ExpandSummaryResponse{
		WhyIds:   []string{"why-1"},
		HowIds:   []string{"how-1", "how-2"},
		HopTargets: []string{"hop-1"},
		Connected: []*ipcv1.ConnectedRecord{
			{RecordId: "conn-1", Text: "connected record 1", Depth: 1, EdgeWeight: 0.5, EdgeType: "how_ids"},
		},
	}, nil
}

func (f *FakeBackend) DaemonStatus(ctx context.Context, req *ipcv1.DaemonStatusRequest) (*ipcv1.DaemonStatusResponse, error) {
	return &ipcv1.DaemonStatusResponse{
		Ok: true, Version: "test", Uptime: "1h",
		CurrentOpenTenants: 1, MaxOpenTenants: 10, CacheEntries: 100,
		CacheSize: 1024, CacheMaxSize: 2048, CacheHitRate: 0.75, GlobalDbHealthy: true,
	}, nil
}

// NewFakeBackend creates an in-process gRPC server with bufconn.
// Returns the fake backend (for setting up test data), a client, and a cleanup function.
func NewFakeBackend(t *testing.T) (*FakeBackend, ipcv1.LibravDBClient, func()) {
	t.Helper()

	fake := &FakeBackend{HealthOK: true}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	ipcv1.RegisterLibravDBServer(server, fake)

	go func() {
		_ = server.Serve(listener)
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Stop()
	}

	return fake, ipcv1.NewLibravDBClient(conn), cleanup
}
