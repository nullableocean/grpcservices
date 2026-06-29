package xrequestid

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type xreqContextKey struct{}

const (
	XREQUEST_ID_KEY = "x-request-id"
)

// GetFromIncomingCtx извлекает из контекста x-request-id
//
// "" если не найден
func GetFromIncomingCtx(ctx context.Context) string {
	meta, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		v := ctx.Value(xreqContextKey{})
		key, ok := v.(string)
		if ok {
			return key
		}

		return ""
	}

	val := meta.Get(XREQUEST_ID_KEY)
	if len(val) == 0 {
		return ""
	}

	return val[0]
}

// CreateNewToOutCtx генерирует x-request-id и записывает в исходящую метадату и в значение контекста
func CreateNewToOutCtx(ctx context.Context) context.Context {
	xrequestId := NewXRequestId()
	return SetToOutCtx(xrequestId, ctx)
}

// SetToOutCtx записывает x-request-id в исходящую метадату и в значение контекста
func SetToOutCtx(xreqid string, ctx context.Context) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, XREQUEST_ID_KEY, xreqid)
	return context.WithValue(ctx, xreqContextKey{}, xreqid)
}
