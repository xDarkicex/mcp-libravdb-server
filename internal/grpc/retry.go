package grpc

import (
	"context"
	"math/rand/v2"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryInterceptor returns a unary client interceptor that retries on Unavailable and DeadlineExceeded.
// 3 attempts, 500ms base, jittered exponential backoff. Matches OpenClaw plugin retry behavior.
func RetryInterceptor(maxRetries int, baseDelay time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				if err := sleepJittered(ctx, baseDelay, attempt); err != nil {
					return err
				}
			}
			lastErr = invoker(ctx, method, req, reply, cc, opts...)
			if lastErr == nil || !isRetryable(lastErr) {
				return lastErr
			}
		}
		return lastErr
	}
}

func sleepJittered(ctx context.Context, base time.Duration, attempt int) error {
	jitter := time.Duration(rand.Int64N(int64(base)))
	delay := base*time.Duration(1<<(attempt-1)) + jitter
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

func isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded
}
