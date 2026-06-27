package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xDarkicex/pidpeek"

	"github.com/xDarkicex/MCP-memory-libravdb/internal/app"
	"github.com/xDarkicex/MCP-memory-libravdb/internal/grpc"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status and tenant list",
	Long:  "Show daemon health, version, tenant list, and optionally MCP server process resource utilization.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg := configFromViper()

		conn, client, err := grpc.Dial(cfg.BackendAddr, cfg.BackendTLS, cfg.BackendTimeout, cfg.ResolveTenantKey())
		if err != nil {
			return fmt.Errorf("daemon connection failed: %w", err)
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.BackendTimeout)
		defer cancel()

		resp, err := client.DaemonStatus(ctx, &ipcv1.DaemonStatusRequest{})
		if err != nil {
			return fmt.Errorf("daemon status query failed: %w", err)
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if jsonOut {
			return printStatusJSON(resp, verbose)
		}
		printStatusText(resp, verbose, cfg)
		return nil
	},
}

var workspacesCmd = &cobra.Command{
	Use:   "workspaces",
	Short: "List all tenant workspaces",
	Long:  "List all tenant workspaces registered with the daemon. Use these keys to grant read access to chat agents in workspace.yaml.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg := configFromViper()

		conn, client, err := grpc.Dial(cfg.BackendAddr, cfg.BackendTLS, cfg.BackendTimeout, cfg.ResolveTenantKey())
		if err != nil {
			return fmt.Errorf("daemon connection failed: %w", err)
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.BackendTimeout)
		defer cancel()

		resp, err := client.DaemonStatus(ctx, &ipcv1.DaemonStatusRequest{})
		if err != nil {
			return fmt.Errorf("daemon status query failed: %w", err)
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			tenants := make([]map[string]interface{}, len(resp.Tenants))
			for i, t := range resp.Tenants {
				tenants[i] = map[string]interface{}{
					"key":           t.TenantKey,
					"status":        t.Status,
					"size_bytes":    t.SizeBytes,
					"open_sessions": t.OpenSessions,
					"last_active":   t.LastActiveMs,
				}
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tenants)
		}

		fmt.Printf("Tenants on %s\n\n", cfg.BackendAddr)
		fmt.Printf("%-55s %-12s %10s %10s\n", "TENANT KEY", "STATUS", "SIZE", "SESSIONS")
		fmt.Println("─────────────────────────────────────────────────────── ──────────── ────────── ──────────")
		for _, t := range resp.Tenants {
			status := t.Status
			if t.Unregistered {
				status = "unregistered"
			}
			fmt.Printf("%-55s %-12s %10s %10d\n",
				t.TenantKey, status,
				pidpeek.FormatBytes(uint64(t.SizeBytes)),
				t.OpenSessions,
			)
		}
		if len(resp.Tenants) == 0 {
			fmt.Println("  (no tenants)")
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolP("verbose", "v", false, "Include MCP server process resource utilization (RSS, VMS, heap, goroutines)")
	statusCmd.Flags().Bool("json", false, "Machine-readable JSON output")
	rootCmd.AddCommand(statusCmd)

	workspacesCmd.Flags().Bool("json", false, "Machine-readable JSON output")
	rootCmd.AddCommand(workspacesCmd)
}

func printStatusText(resp *ipcv1.DaemonStatusResponse, verbose bool, cfg *app.Config) {
	fmt.Printf("libravdbd %s  uptime=%s\n", resp.Version, resp.Uptime)
	fmt.Printf("  Backend:   %s  (embed-init: %.1fs)\n", resp.Backend, resp.EmbedInitSecs)

	healthy := "[healthy]"
	if !resp.GlobalDbHealthy {
		healthy = "[unhealthy]"
	}
	fmt.Printf("\nControl Plane: daemon.libravdb  (%s)  %s\n",
		pidpeek.FormatBytes(uint64(resp.GlobalDbSize)), healthy)

	fmt.Printf("\nWorkloads: %d/%d tenants  |  Cache: %d entries, %s / %s (%.1f%% hit)\n",
		resp.CurrentOpenTenants, resp.MaxOpenTenants,
		resp.CacheEntries,
		pidpeek.FormatBytes(uint64(resp.CacheSize)),
		pidpeek.FormatBytes(uint64(resp.CacheMaxSize)),
		resp.CacheHitRate*100,
	)

	fmt.Println()
	fmt.Printf("%-55s %-12s %10s %10s\n", "TENANT KEY", "STATUS", "SIZE", "SESSIONS")
	fmt.Println("─────────────────────────────────────────────────────── ──────────── ────────── ──────────")
	for _, t := range resp.Tenants {
		status := t.Status
		if t.Unregistered {
			status = "unregistered"
		}
		fmt.Printf("%-55s %-12s %10s %10d\n",
			t.TenantKey, status,
			pidpeek.FormatBytes(uint64(t.SizeBytes)),
			t.OpenSessions,
		)
	}

	fmt.Println()
	fmt.Printf("MCP server:  tenant=%s  addr=%s\n", cfg.ResolveTenantKey(), cfg.BackendAddr)

	if verbose {
		fmt.Println()
		fmt.Println("MCP Server Process:")
		fmt.Printf("  Heap:       %s\n", pidpeek.FormatBytes(pidpeek.GoHeapAlloc()))
		fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())

		pex, err := pidpeek.GetAll(os.Getpid())
		if err == nil {
			fmt.Printf("  RSS:        %s\n", pidpeek.FormatBytes(pex.RSS))
			fmt.Printf("  VMS:        %s\n", pidpeek.FormatBytes(pex.VMSSize))
			fmt.Printf("  Threads:    %d\n", pex.ThreadNum)
			fmt.Printf("  CPU:        %.1fs\n", pex.CPUTotalSec)
		}
	}
}

func printStatusJSON(resp *ipcv1.DaemonStatusResponse, verbose bool) error {
	out := map[string]interface{}{
		"version":           resp.Version,
		"uptime":            resp.Uptime,
		"backend":           resp.Backend,
		"embed_init_secs":   resp.EmbedInitSecs,
		"global_db_healthy": resp.GlobalDbHealthy,
		"global_db_size":    resp.GlobalDbSize,
		"open_tenants":      resp.CurrentOpenTenants,
		"max_tenants":       resp.MaxOpenTenants,
		"cache_entries":     resp.CacheEntries,
		"cache_size":        resp.CacheSize,
		"cache_max_size":    resp.CacheMaxSize,
		"cache_hit_rate":    resp.CacheHitRate,
	}

	tenants := make([]map[string]interface{}, len(resp.Tenants))
	for i, t := range resp.Tenants {
		status := t.Status
		if t.Unregistered {
			status = "unregistered"
		}
		tenants[i] = map[string]interface{}{
			"key":           t.TenantKey,
			"status":        status,
			"size_bytes":    t.SizeBytes,
			"open_sessions": t.OpenSessions,
		}
	}
	out["tenants"] = tenants

	if verbose {
		out["mcp_server"] = map[string]interface{}{
			"goroutines":    runtime.NumGoroutine(),
			"go_heap_bytes": pidpeek.GoHeapAlloc(),
		}
		if pex, err := pidpeek.GetAll(os.Getpid()); err == nil {
			out["mcp_server"].(map[string]interface{})["rss_bytes"] = pex.RSS
			out["mcp_server"].(map[string]interface{})["vms_bytes"] = pex.VMSSize
			out["mcp_server"].(map[string]interface{})["threads"] = pex.ThreadNum
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
