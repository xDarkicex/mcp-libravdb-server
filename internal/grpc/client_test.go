package grpc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestParseAddr_Unix(t *testing.T) {
	target, dialer := parseAddr("unix:///tmp/test.sock")
	if target != "/tmp/test.sock" {
		t.Fatalf("expected /tmp/test.sock, got %s", target)
	}
	conn, err := dialer(context.Background(), target)
	if err == nil {
		conn.Close()
	}
}

func TestParseAddr_UnixHomeDir(t *testing.T) {
	target, _ := parseAddr("unix://~/.libravdbd/run/libravdb.sock")
	if strings.HasPrefix(target, "~") {
		t.Fatal("expected home dir to be expanded")
	}
	if !strings.HasSuffix(target, ".libravdbd/run/libravdb.sock") {
		t.Fatalf("unexpected path: %s", target)
	}
}

func TestParseAddr_TCP(t *testing.T) {
	target, _ := parseAddr("127.0.0.1:7654")
	if target != "127.0.0.1:7654" {
		t.Fatalf("expected 127.0.0.1:7654, got %s", target)
	}
}

func TestResolveAuthSecret_Env(t *testing.T) {
	os.Setenv("LIBRAVDB_AUTH_SECRET", "test-secret")
	defer os.Unsetenv("LIBRAVDB_AUTH_SECRET")

	secret := resolveAuthSecret()
	if secret != "test-secret" {
		t.Fatalf("expected test-secret, got %s", secret)
	}
}

func TestResolveAuthSecret_File(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")

	tmp := filepath.Join(t.TempDir(), "secret.txt")
	os.WriteFile(tmp, []byte("file-secret\n"), 0600)
	os.Setenv("LIBRAVDB_AUTH_SECRET_FILE", tmp)
	defer os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	secret := resolveAuthSecret()
	if secret != "file-secret" {
		t.Fatalf("expected file-secret, got %q", secret)
	}
}

func TestResolveAuthSecret_None(t *testing.T) {
	os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	secret := resolveAuthSecret()
	if secret != "" {
		t.Fatalf("expected empty, got %q", secret)
	}
}

func TestTenantKeyInterceptor(t *testing.T) {
	interceptor := tenantKeyInterceptor("test-tenant")
	err := interceptor(context.Background(), "/test.Method", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
