package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xDarkicex/pidpeek"

	"github.com/xDarkicex/mcp-libravdb-server/internal/grpc"
	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage libravdb tenants (workspaces)",
	Long:  "List, evict, or inspect tenant workspaces. Tenants are auto-created on first use — no registration needed.",
}

var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all open tenants",
	Long:  "List all currently open tenant workspaces. Uses DaemonStatus RPC so only shows tenants loaded in memory.",
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
			return fmt.Errorf("status query failed: %w", err)
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			tenants := make([]map[string]interface{}, len(resp.Tenants))
			for i, t := range resp.Tenants {
				s := t.Status
				if t.Unregistered {
					s = "unregistered"
				}
				tenants[i] = map[string]interface{}{
					"key":           t.TenantKey,
					"status":        s,
					"size_bytes":    t.SizeBytes,
					"open_sessions": t.OpenSessions,
					"last_active_ms": t.LastActiveMs,
				}
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tenants)
		}

		fmt.Printf("Open tenants on %s\n\n", cfg.BackendAddr)
		fmt.Printf("%-55s %-12s %10s %10s\n", "TENANT KEY", "STATUS", "SIZE", "SESSIONS")
		fmt.Println("─────────────────────────────────────────────────────── ──────────── ────────── ──────────")
		for _, t := range resp.Tenants {
			s := t.Status
			if t.Unregistered {
				s = "unregistered"
			}
			fmt.Printf("%-55s %-12s %10s %10d\n",
				t.TenantKey, s,
				pidpeek.FormatBytes(uint64(t.SizeBytes)),
				t.OpenSessions,
			)
		}
		if len(resp.Tenants) == 0 {
			fmt.Println("  (no open tenants)")
		}
		return nil
	},
}

var tenantEvictCmd = &cobra.Command{
	Use:   "evict <tenant-key>",
	Short: "Evict a tenant from daemon memory",
	Long:  "Close and unload a tenant from the daemon's memory. The tenant's data remains on disk and will reload on next access.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantKey := args[0]
		cfg := configFromViper()

		conn, client, err := grpc.Dial(cfg.BackendAddr, cfg.BackendTLS, cfg.BackendTimeout, cfg.ResolveTenantKey())
		if err != nil {
			return fmt.Errorf("daemon connection failed: %w", err)
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.BackendTimeout)
		defer cancel()

		resp, err := client.EvictTenant(ctx, &ipcv1.EvictTenantRequest{
			TenantKey: tenantKey,
		})
		if err != nil {
			return fmt.Errorf("evict failed: %w", err)
		}

		if resp.Ok {
			fmt.Printf("Tenant '%s' evicted.\n", tenantKey)
		} else {
			fmt.Printf("Eviction returned: %s\n", resp.Message)
		}
		return nil
	},
}

var tenantInspectCmd = &cobra.Command{
	Use:   "inspect <tenant-key>",
	Short: "Show tenant details and status",
	Long:  "Show detailed information about a specific tenant including status, size, sessions, and active status.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantKey := args[0]
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
			return fmt.Errorf("status query failed: %w", err)
		}

		var found *ipcv1.TenantStatus
		for _, t := range resp.Tenants {
			if t.TenantKey == tenantKey {
				found = t
				break
			}
		}

		if found == nil {
			fmt.Printf("Tenant '%s' is not currently open.\n", tenantKey)
			fmt.Println("(Only open tenants are visible — closed tenants may still exist on disk)")
			return nil
		}

		s := found.Status
		if found.Unregistered {
			s = "unregistered"
		}

		fmt.Printf("Tenant:     %s\n", found.TenantKey)
		fmt.Printf("Status:     %s\n", s)
		fmt.Printf("Size:       %s\n", pidpeek.FormatBytes(uint64(found.SizeBytes)))
		fmt.Printf("Sessions:   %d\n", found.OpenSessions)
		if found.LastActiveMs > 0 {
			fmt.Printf("Last active: %d ms ago\n", found.LastActiveMs)
		}
		return nil
	},
}

func init() {
	tenantListCmd.Flags().Bool("json", false, "Machine-readable JSON output")
	tenantCmd.AddCommand(tenantListCmd)
	tenantCmd.AddCommand(tenantEvictCmd)
	tenantCmd.AddCommand(tenantInspectCmd)
	rootCmd.AddCommand(tenantCmd)
}
