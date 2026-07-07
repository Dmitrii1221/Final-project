package grpc

import (
	"context"
	"final-project/internal/auth"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(secret []byte) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}
		tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")
		claims, err := auth.ValidateToken(tokenStr, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = context.WithValue(ctx, auth.ClaimsKey, claims)
		return handler(ctx, req)
	}
}
