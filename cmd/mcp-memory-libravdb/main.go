package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/xDarkicex/mcp-libravdb-server/internal/app"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	rootCmd = &cobra.Command{
		Use:     "mcp-memory-libravdb",
		Short:   "libravdb Memory MCP Server",
		Long:    "MCP server exposing libravdbd's cognitive memory system to any MCP-compatible AI client.",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  "Start an MCP server over standard input/output (for Claude Desktop, Claude Code, Codex, Cursor, etc.).",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := configFromViper()
			if cfg.BackendAddr == "" {
				return errors.New("backend address required: set LIBRAVDB_BACKEND_ADDR or --backend-addr")
			}
			return app.RunStdio(cfg)
		},
	}

	httpCmd = &cobra.Command{
		Use:   "http",
		Short: "Start HTTP server",
		Long:  "Start an MCP server over Streamable HTTP (for network/remote clients).",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := configFromViper()
			if cfg.BackendAddr == "" {
				return errors.New("backend address required: set LIBRAVDB_BACKEND_ADDR or --backend-addr")
			}
			return app.RunHTTP(cfg)
		},
	}
)

func configFromViper() *app.Config {
	addr := viper.GetString("backend-addr")
	if addr == "" {
		addr = "unix://~/.libravdbd/run/libravdb.sock"
	}
	tenantKey := viper.GetString("tenant-key")
	if tenantKey == "" {
		tenantKey = app.DefaultTenantKey
	}

	return &app.Config{
		BackendAddr:    addr,
		BackendTLS:     viper.GetBool("backend-tls"),
		BackendTimeout: viper.GetDuration("backend-timeout"),
		DegradedOk:     viper.GetBool("degraded-ok"),
		LogLevel:       viper.GetString("log-level"),
		TenantKey:      tenantKey,
		Workspace:      viper.GetString("workspace"),
		Shared:         viper.GetBool("shared"),
		AuthToken:      viper.GetString("auth-token"),
		HTTPPort:       viper.GetInt("http-port"),
		HTTPHost:       viper.GetString("http-host"),
		HTTPExpose:     viper.GetBool("http-expose"),
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetGlobalNormalizationFunc(normalizeFlags)

	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")

	// Global flags
	rootCmd.PersistentFlags().String("backend-addr", "", "gRPC backend address (unix:///path/to/sock or host:port)")
	rootCmd.PersistentFlags().Bool("backend-tls", false, "Enable TLS for gRPC")
	rootCmd.PersistentFlags().Duration("backend-timeout", app.DefaultTimeout, "Per-call gRPC timeout")
	rootCmd.PersistentFlags().Bool("degraded-ok", false, "Run degraded if backend unavailable at startup")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug, info, warn, error")
	rootCmd.PersistentFlags().String("tenant-key", app.DefaultTenantKey, "libravdb tenant key for data isolation")
	rootCmd.PersistentFlags().Bool("shared", false, "Disable per-workspace isolation (all workspaces share one tenant)")
	rootCmd.PersistentFlags().String("workspace", "", "Explicit workspace name (appended to tenant key)")
	rootCmd.PersistentFlags().String("auth-token", "", "Bearer token for HTTP authentication (MCP_AUTH_TOKEN)")

	// HTTP-specific flags
	httpCmd.Flags().Int("http-port", 8082, "HTTP server port")
	httpCmd.Flags().String("http-host", "127.0.0.1", "HTTP server bind address")
	httpCmd.Flags().Bool("http-expose", false, "Expose HTTP server on non-localhost interfaces")

	// Bind flags to viper
	_ = viper.BindPFlag("backend-addr", rootCmd.PersistentFlags().Lookup("backend-addr"))
	_ = viper.BindPFlag("backend-tls", rootCmd.PersistentFlags().Lookup("backend-tls"))
	_ = viper.BindPFlag("backend-timeout", rootCmd.PersistentFlags().Lookup("backend-timeout"))
	_ = viper.BindPFlag("degraded-ok", rootCmd.PersistentFlags().Lookup("degraded-ok"))
	_ = viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("tenant-key", rootCmd.PersistentFlags().Lookup("tenant-key"))
	_ = viper.BindPFlag("shared", rootCmd.PersistentFlags().Lookup("shared"))
	_ = viper.BindPFlag("workspace", rootCmd.PersistentFlags().Lookup("workspace"))
	_ = viper.BindPFlag("auth-token", rootCmd.PersistentFlags().Lookup("auth-token"))
	_ = viper.BindPFlag("http-port", httpCmd.Flags().Lookup("http-port"))
	_ = viper.BindPFlag("http-host", httpCmd.Flags().Lookup("http-host"))
	_ = viper.BindPFlag("http-expose", httpCmd.Flags().Lookup("http-expose"))

	rootCmd.AddCommand(stdioCmd)
	rootCmd.AddCommand(httpCmd)
}

func initConfig() {
	viper.SetEnvPrefix("libravdb")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	_ = viper.BindEnv("tenant-key", "LIBRAVDB_TENANT_KEY")
	_ = viper.BindEnv("auth-token", "LIBRAVDB_AUTH_TOKEN")
}

func normalizeFlags(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
