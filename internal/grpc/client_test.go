package grpc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

func TestParseAddr_Unix(t *testing.T) {
	target, dialer := parseAddr("unix:///tmp/test.sock")
	if target != "/tmp/test.sock" {
		t.Fatalf("expected /tmp/test.sock, got %s", target)
	}
	conn, err := dialer(context.Background(), target)
	if err == nil {
		_ = conn.Close()
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
	_ = os.Setenv("LIBRAVDB_AUTH_SECRET", "test-secret")
	t.Cleanup(func() { _ = os.Unsetenv("LIBRAVDB_AUTH_SECRET") })

	secret := resolveAuthSecret()
	if secret != "test-secret" {
		t.Fatalf("expected test-secret, got %s", secret)
	}
}

func TestResolveAuthSecret_File(t *testing.T) {
	_ = os.Unsetenv("LIBRAVDB_AUTH_SECRET")

	tmp := filepath.Join(t.TempDir(), "secret.txt")
	_ = os.WriteFile(tmp, []byte("file-secret\n"), 0600)
	_ = os.Setenv("LIBRAVDB_AUTH_SECRET_FILE", tmp)
	t.Cleanup(func() { _ = os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE") })

	secret := resolveAuthSecret()
	if secret != "file-secret" {
		t.Fatalf("expected file-secret, got %q", secret)
	}
}

func TestResolveAuthSecret_None(t *testing.T) {
	_ = os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	_ = os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	secret := resolveAuthSecret()
	if secret != "" {
		t.Fatalf("expected empty, got %q", secret)
	}
}

func TestTenantKeyInterceptor(t *testing.T) {
	interceptor := tenantKeyInterceptor("test-tenant")
	err := interceptor(context.Background(), "/test.Method", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected metadata in context")
			}
			if md.Get("libravdb-tenant-key")[0] != "test-tenant" {
				t.Fatal("tenant key not set")
			}
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHMACAuthInterceptor(t *testing.T) {
	_ = os.Setenv("LIBRAVDB_AUTH_SECRET", "test-secret")
	t.Cleanup(func() { _ = os.Unsetenv("LIBRAVDB_AUTH_SECRET") })

	interceptor := hmacAuthInterceptor("test-secret")
	err := interceptor(context.Background(), "/test.Method", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected metadata")
			}
			if len(md.Get("x-libravdb-nonce")) == 0 {
				t.Fatal("nonce not set")
			}
			if len(md.Get("x-libravdb-auth")) == 0 {
				t.Fatal("auth signature not set")
			}
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDial_Bufconn(t *testing.T) {
	_ = os.Unsetenv("LIBRAVDB_AUTH_SECRET")
	_ = os.Unsetenv("LIBRAVDB_AUTH_SECRET_FILE")

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	ipcv1.RegisterLibravDBServer(server, &ipcv1.UnimplementedLibravDBServer{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()

	conn, client, err := Dial("passthrough:///bufnet", false, 5*time.Second, "test-tenant",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil conn")
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	_ = conn.Close()
}

func TestDial_TLSNotImplemented(t *testing.T) {
	_, _, err := Dial("unix:///tmp/test.sock", true, 5*time.Second, "test")
	if err == nil {
		t.Fatal("expected TLS error")
	}
}

func TestClientConn_Close(t *testing.T) {
	cc := &ClientConn{ClientConn: nil}
	if err := cc.Close(); err != nil {
		t.Fatal("Close on nil should not error")
	}
}
