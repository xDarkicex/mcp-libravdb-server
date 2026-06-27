package app

import (
	"os"
	"testing"
)

func TestResolveTenantKey_DefaultIsolation(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := &Config{TenantKey: DefaultTenantKey}

	key := cfg.ResolveTenantKey()
	if len(key) <= len(DefaultTenantKey)+1 {
		t.Fatalf("expected isolated key like libravdb-mcp-server:<hash>, got %s", key)
	}
	if key[:len(DefaultTenantKey)] != DefaultTenantKey {
		t.Fatalf("expected prefix %s, got %s", DefaultTenantKey, key)
	}
	_ = cwd // used by hashCWD
}

func TestResolveTenantKey_Shared(t *testing.T) {
	cfg := &Config{TenantKey: DefaultTenantKey, Shared: true}

	key := cfg.ResolveTenantKey()
	if key != DefaultTenantKey {
		t.Fatalf("expected %s, got %s", DefaultTenantKey, key)
	}
}

func TestResolveTenantKey_Workspace(t *testing.T) {
	cfg := &Config{TenantKey: DefaultTenantKey, Workspace: "myproj"}

	key := cfg.ResolveTenantKey()
	expected := DefaultTenantKey + ":myproj"
	if key != expected {
		t.Fatalf("expected %s, got %s", expected, key)
	}
}

func TestResolveTenantKey_CustomShared(t *testing.T) {
	cfg := &Config{TenantKey: "custom-key", Shared: true}

	key := cfg.ResolveTenantKey()
	if key != "custom-key" {
		t.Fatalf("expected custom-key, got %s", key)
	}
}

func TestHashCWD(t *testing.T) {
	h := hashCWD()
	if len(h) != 8 {
		t.Fatalf("expected 8-char hex hash, got %s", h)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected int // slog.Level is int
	}{
		{"debug", -4},
		{"info", 0},
		{"warn", 4},
		{"error", 8},
		{"unknown", 0}, // default to info
		{"", 0},
	}
	for _, tt := range tests {
		got := parseLogLevel(tt.input)
		if int(got) != tt.expected {
			t.Errorf("parseLogLevel(%q) = %d, want %d", tt.input, int(got), tt.expected)
		}
	}
}
