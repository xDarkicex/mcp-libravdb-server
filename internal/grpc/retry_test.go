package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code     codes.Code
		retryable bool
	}{
		{codes.Unavailable, true},
		{codes.DeadlineExceeded, true},
		{codes.Internal, false},
		{codes.InvalidArgument, false},
		{codes.OK, false},
	}
	for _, tt := range tests {
		err := status.Error(tt.code, "test")
		if got := isRetryable(err); got != tt.retryable {
			t.Errorf("isRetryable(%s) = %v, want %v", tt.code, got, tt.retryable)
		}
	}

	// Non-gRPC errors are not retryable
	if isRetryable(errors.New("plain error")) {
		t.Error("plain error should not be retryable")
	}
}

func TestSleepJittered_Success(t *testing.T) {
	ctx := context.Background()
	err := sleepJittered(ctx, 1*time.Millisecond, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSleepJittered_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepJittered(ctx, 1*time.Second, 1)
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestRetryInterceptor_Success(t *testing.T) {
	interceptor := RetryInterceptor(2, 1*time.Millisecond)
	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryInterceptor_NonRetryable(t *testing.T) {
	interceptor := RetryInterceptor(2, 1*time.Millisecond)
	calls := 0
	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			calls++
			return status.Error(codes.Internal, "not retryable")
		})
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRetryInterceptor_RetryableEventuallySucceeds(t *testing.T) {
	interceptor := RetryInterceptor(3, 1*time.Millisecond)
	calls := 0
	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			calls++
			if calls < 2 {
				return status.Error(codes.Unavailable, "transient")
			}
			return nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestRetryInterceptor_Exhausted(t *testing.T) {
	interceptor := RetryInterceptor(1, 1*time.Millisecond)
	calls := 0
	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			calls++
			return status.Error(codes.Unavailable, "always down")
		})
	if calls != 2 { // initial + 1 retry
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
}
