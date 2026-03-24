package intercepter

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nullableocean/grpcservices/shared/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	TokenMetadataKey = "authorization"

	UserClaimKey    = "sub"
	RolesClaimKey   = "rls"
	ExpiredClaimKey = "exp"

	UserCtxKey  = "user_uuid"
	RolesCtxKey = "roles"
	TokenCtxKey = "token"
)

type JwtAuthorizer interface {
	ParseToken(t string) (*jwt.Token, error)
	ExtractClaims(t *jwt.Token) (map[string]interface{}, error)
}

func UnaryJwtAuthInterceptor(logger *zap.Logger, auth auth.JwtAuthorizer) grpc.UnaryServerInterceptor {
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

func StreamJwtAuthInterceptor(logger *zap.Logger, auth JwtAuthorizer) grpc.StreamServerInterceptor {
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

func parseJwtAndInjectInCtx(ctx context.Context, logger *zap.Logger, auth JwtAuthorizer) (context.Context, error) {
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

	jwtTkn, err := auth.ParseToken(token[0])
	if err != nil {
		logger.Warn("failed parse jwt", zap.Error(err))

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	claims, err := auth.ExtractClaims(jwtTkn)
	if err != nil {
		logger.Warn("failed extract jwt claims", zap.Error(err))

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := validateExpired(claims); err != nil {
		return nil, err
	}

	userUUID, err := getUserUUID(claims)
	if err != nil {
		logger.Warn("failed extract user uuid from jwt claims", zap.Error(err))

		return nil, err
	}

	roles, err := getUserRoles(claims)
	if err != nil {
		logger.Warn("failed extract roles from jwt claims")

		return nil, err
	}

	newCtx := context.WithValue(ctx, UserCtxKey, userUUID)
	newCtx = context.WithValue(newCtx, RolesCtxKey, roles)
	newCtx = context.WithValue(newCtx, TokenCtxKey, token[0])

	return newCtx, nil
}

// getUserUUID find User UUID in JWT Claims by "sub" key
func getUserUUID(claims map[string]interface{}) (string, error) {
	userUUID, ok := claims[UserClaimKey].(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user uuid not provided in jwt")
	}

	return userUUID, nil
}

func getUserRoles(claims map[string]interface{}) ([]string, error) {
	rawRoles, ok := claims[RolesClaimKey].([]interface{})
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user roles not provided in jwt")
	}

	roles := make([]string, 0, len(rawRoles))
	for _, r := range rawRoles {
		role, ok := r.(string)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "non-string role in roles array")
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// validateExpired validate expires jwt by "exp" key
func validateExpired(claims map[string]interface{}) error {
	exp, ok := claims[ExpiredClaimKey].(float64)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid exp")
	}

	if time.Now().After(time.Unix(int64(exp), 0)) {
		return status.Error(codes.Unauthenticated, "expired token")
	}

	return nil
}

// UnaryClientJwtForwardInterceptor
// setup outgoing context with auth token
func UnaryClientJwtForwardInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token, ok := ctx.Value(TokenCtxKey).(string)
		if ok && token != "" {
			md, ex := metadata.FromOutgoingContext(ctx)
			if !ex {
				md = metadata.New(nil)
			} else {
				md = md.Copy()
			}

			md.Set(TokenMetadataKey, token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
