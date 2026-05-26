package interceptors

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const (
	TokenMetadataKey = "authorization"

	userCtxKey  ctxKey = "user_uuid"
	rolesCtxKey ctxKey = "roles"
	tokenCtxKey ctxKey = "token"
)

func UserUUIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userCtxKey).(string)
	return v, ok
}

func RolesFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(rolesCtxKey).([]string)
	return v, ok
}

func JwtTokenFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(tokenCtxKey).(string)
	return v, ok
}

func UnaryJwtAuthInterceptor(logger *zap.Logger, auth auth.JwtParser) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		newCtx, err := parseJwtAndInjectInCtx(ctx, logger, auth)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func StreamJwtAuthInterceptor(logger *zap.Logger, auth auth.JwtParser) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		newCtx, err := parseJwtAndInjectInCtx(ctx, logger, auth)
		if err != nil {
			return err
		}

		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

func parseJwtAndInjectInCtx(ctx context.Context, logger *zap.Logger, auth auth.JwtParser) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Warn("failed jwt auth: metadata not found")

		return nil, status.Error(codes.Unauthenticated, "metadata not found")
	}

	token := md.Get(TokenMetadataKey)
	if len(token) == 0 {
		logger.Warn("failed jwt auth: auth token is not found")

		return nil, status.Error(codes.Unauthenticated, "auth token is not found")
	}

	userClaims, err := auth.ParseToken(token[0])
	if err != nil {
		logger.Warn("failed parse user claims from jwt", zap.Error(err))

		return nil, status.Error(codes.Unauthenticated, "failed jwt")
	}

	userUUID := userClaims.GetUserUUID()
	roles := userClaims.GetRoles()

	newCtx := context.WithValue(ctx, userCtxKey, userUUID)
	newCtx = context.WithValue(newCtx, rolesCtxKey, roles)
	newCtx = context.WithValue(newCtx, tokenCtxKey, token[0])

	return newCtx, nil
}

// UnaryClientJwtForwardInterceptor
// setup outgoing context with auth token
func UnaryClientJwtForwardInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token, ok := ctx.Value(tokenCtxKey).(string)
		if ok && token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, TokenMetadataKey, token)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
