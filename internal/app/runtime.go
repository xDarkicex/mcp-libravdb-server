package app

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xDarkicex/nanite"
	nanitesse "github.com/xDarkicex/nanite/sse"

	"github.com/xDarkicex/MCP-memory-libravdb/internal/grpc"
	"github.com/xDarkicex/MCP-memory-libravdb/internal/resources"
	"github.com/xDarkicex/MCP-memory-libravdb/internal/tools"
	"github.com/xDarkicex/MCP-memory-libravdb/internal/transport"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

type backendConnection struct {
	conn    *grpc.ClientConn
	client  ipcv1.LibravDBClient
	healthy bool
}

// Runtime holds the wired server dependencies.
type Runtime struct {
	Config     *Config
	Logger     *slog.Logger
	MCP        *mcp.Server
	gRPCClient ipcv1.LibravDBClient
	gRPCConn   *grpc.ClientConn
}

func connectBackend(cfg *Config, logger *slog.Logger) (backendConnection, error) {
	conn, client, err := grpc.Dial(cfg.BackendAddr, cfg.BackendTLS, cfg.BackendTimeout, cfg.ResolveTenantKey())
	if err != nil {
		if cfg.DegradedOk {
			logger.Warn("backend unavailable, starting in degraded mode", "addr", cfg.BackendAddr, "err", err)
			return backendConnection{}, nil
		}
		return backendConnection{}, fmt.Errorf("backend connection failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.BackendTimeout)
	defer cancel()

	resp, err := client.Health(ctx, &ipcv1.HealthRequest{})
	if err != nil || !resp.Ok {
		if cfg.DegradedOk {
			logger.Warn("backend health check failed, starting degraded", "err", err)
			return backendConnection{conn: conn, client: client}, nil
		}
		return backendConnection{}, fmt.Errorf("backend health check failed: %w", err)
	}

	logger.Info("backend healthy", "addr", cfg.BackendAddr, "tenant", cfg.ResolveTenantKey())
	return backendConnection{conn: conn, client: client, healthy: true}, nil
}

// NewRuntime creates and wires all server dependencies.
func NewRuntime(cfg *Config) (*Runtime, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))

	be, err := connectBackend(cfg, logger)
	if err != nil {
		return nil, err
	}

	mcpServer := NewMCP(logger, be.healthy)

	deps := &tools.Deps{
		Client:         be.client,
		Logger:         logger,
		BackendTimeout: cfg.BackendTimeout,
		BackendHealthy: be.healthy,
	}
	if err := tools.RegisterAll(mcpServer, deps); err != nil {
		return nil, fmt.Errorf("tool registration failed: %w", err)
	}
	resources.RegisterAll(mcpServer)

	return &Runtime{
		Config:     cfg,
		Logger:     logger,
		MCP:        mcpServer,
		gRPCClient: be.client,
		gRPCConn:   be.conn,
	}, nil
}

// Shutdown gracefully closes connections.
func (r *Runtime) Shutdown() {
	if r.gRPCConn != nil {
		if err := r.gRPCConn.Close(); err != nil {
			r.Logger.Warn("gRPC connection close error", "err", err)
		}
	}
	r.Logger.Info("server shutdown complete")
}

// RunStdio starts the MCP server over stdio transport.
func RunStdio(cfg *Config) error {
	rt, err := NewRuntime(cfg)
	if err != nil {
		return err
	}
	defer rt.Shutdown()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rt.Logger.Info("starting stdio MCP server", "version", "1.0.0")
	return rt.MCP.Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP starts the MCP server over Streamable HTTP with nanite/sse backing the SSE path.
func RunHTTP(cfg *Config) error {
	rt, err := NewRuntime(cfg)
	if err != nil {
		return err
	}
	defer rt.Shutdown()

	r := setupRouter(rt, cfg)
	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	rt.Logger.Info("starting HTTP MCP server", "addr", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		rt.Logger.Info("shutting down HTTP server")
		if err := r.Shutdown(5 * time.Second); err != nil {
			rt.Logger.Warn("HTTP shutdown error", "err", err)
		}
	}()

	if err := r.Start(addr); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

func setupRouter(rt *Runtime, cfg *Config) *nanite.Router {
	r := nanite.New(
		nanite.WithPanicRecovery(true),
		nanite.WithServerTimeouts(5*time.Second, 60*time.Second, 60*time.Second),
	)
	r.Use(requestLogger(rt.Logger))
	r.Use(nanite.CORSMiddleware(nil))

	if cfg.HTTPExpose && cfg.AuthToken != "" {
		r.Use(AuthMiddleware(cfg.AuthToken))
	}

	streamableHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return rt.MCP }, nil,
	)
	r.Post("/mcp", func(c *nanite.Context) {
		streamableHandler.ServeHTTP(c.Writer, c.Request)
	})

	registerSSE(r, rt)

	r.Get("/health", func(c *nanite.Context) {
		c.JSON(200, map[string]interface{}{"backend_healthy": rt.gRPCConn != nil})
	})
	return r
}

func registerSSE(r *nanite.Router, rt *Runtime) {
	nanitesse.Register(r, "/mcp", func(conn *nanitesse.Connection, c *nanite.Context) {
		w := transport.NewSSEWriter(conn)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		sessionID := rand.Text()
		sseT := &mcp.SSEServerTransport{Endpoint: "/mcp?sessionid=" + sessionID, Response: w}

		ss, err := rt.MCP.Connect(c.Request.Context(), sseT, nil)
		if err != nil {
			rt.Logger.Error("SSE connection failed", "err", err)
			return
		}
		defer ss.Close()
		<-c.Request.Context().Done()
	})
}

func requestLogger(logger *slog.Logger) func(*nanite.Context, func()) {
	return func(c *nanite.Context, next func()) {
		start := time.Now()
		next()
		logger.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.GetStatus(),
			"duration", time.Since(start).String(),
		)
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
