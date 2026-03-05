package middleware

import (
	"context"

	"github.com/example/go-rate-limiter/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryRateLimitInterceptor(svc *service.RateLimiterService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		identifier := "anonymous"
		scope := service.ScopeGlobal
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-user-id"); len(vals) > 0 {
				identifier = vals[0]
				scope = service.ScopeUser
			}
		}
		allowed, _, err := svc.Allow(ctx, service.Request{Scope: scope, Identifier: identifier, Endpoint: info.FullMethod})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !allowed {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}
