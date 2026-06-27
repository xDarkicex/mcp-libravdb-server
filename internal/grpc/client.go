package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	ipcv1 "github.com/xDarkicex/libravdb-contracts/libravdb/ipc/v1"
)

// ClientConn wraps a gRPC connection and client.
type ClientConn struct {
	*grpc.ClientConn
}

// Close closes the underlying gRPC connection.
func (c *ClientConn) Close() error {
	if c.ClientConn != nil {
		return c.ClientConn.Close()
	}
	return nil
}

// Dial connects to libravdbd and returns a client with auth and tenant interceptors.
func Dial(addr string, tlsEnabled bool, timeout time.Duration, tenantKey string) (*ClientConn, ipcv1.LibravDBClient, error) {
	target, dialer := parseAddr(addr)

	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return dialer(ctx, target)
		}),
	}

	if tlsEnabled {
		return nil, nil, fmt.Errorf("TLS not yet implemented — use Unix socket")
	}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Auth interceptor (optional, only when secret is configured)
	secret := resolveAuthSecret()
	if secret != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(hmacAuthInterceptor(secret)))
	}

	// Tenant key interceptor (always present)
	opts = append(opts, grpc.WithUnaryInterceptor(tenantKeyInterceptor(tenantKey)))

	// Retry interceptor: 3 attempts, 500ms base jittered backoff on transient errors
	opts = append(opts, grpc.WithUnaryInterceptor(RetryInterceptor(3, 500*time.Millisecond)))

	conn, err := grpc.NewClient("passthrough:///"+target, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	client := ipcv1.NewLibravDBClient(conn)
	return &ClientConn{conn}, client, nil
}

func resolveAuthSecret() string {
	if s := os.Getenv("LIBRAVDB_AUTH_SECRET"); s != "" {
		return s
	}
	if f := os.Getenv("LIBRAVDB_AUTH_SECRET_FILE"); f != "" {
		data, err := os.ReadFile(f)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

func parseAddr(addr string) (string, func(context.Context, string) (net.Conn, error)) {
	if strings.HasPrefix(addr, "unix://") {
		path := strings.TrimPrefix(addr, "unix://")
		if strings.HasPrefix(path, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				path = filepath.Join(home, path[1:])
			}
		}
		return path, func(ctx context.Context, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", path)
		}
	}

	return addr, func(ctx context.Context, target string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "tcp", target)
	}
}

func hmacAuthInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		nonce := fmt.Sprintf("%d", time.Now().UnixNano())
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(nonce + ":" + method))
		sig := hex.EncodeToString(mac.Sum(nil))

		ctx = metadata.AppendToOutgoingContext(ctx,
			"x-libravdb-nonce", nonce,
			"x-libravdb-auth", sig,
		)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func tenantKeyInterceptor(tenantKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		ctx = metadata.AppendToOutgoingContext(ctx, "libravdb-tenant-key", tenantKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
