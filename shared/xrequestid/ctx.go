package xrequestid

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	XREQUEST_ID_KEY = "x-request-id"
)

// GetFromIncomingCtx извлекает из контекста x-request-id
//
// "" если не найден
func GetFromIncomingCtx(ctx context.Context) string {
	meta, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		return ""
	}

	val := meta.Get(XREQUEST_ID_KEY)
	if len(val) == 0 {
		return ""
	}

	return val[0]
}

// CreateToOutCtx генерирует x-request-id и записывает в исходящий контекст
func CreateToOutCtx(ctx context.Context) context.Context {
	xrequestId := NewXRequestId()
	return SetInOutCtx(xrequestId, ctx)
}

func SetInOutCtx(xreqid string, ctx context.Context) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, XREQUEST_ID_KEY, xreqid)
	return ctx
}
