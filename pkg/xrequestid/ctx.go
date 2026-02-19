package xrequestid

import (
	"context"

	"google.golang.org/grpc/metadata"
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

// NewForOutCtx генерирует x-request-id и записывает в исходящий контекст
func NewForOutCtx(ctx context.Context) context.Context {
	xrequestId := NewXRequestId()
	return SetInOutCtx(xrequestId, ctx)
}

func SetInOutCtx(xreqid string, ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{"x-request-id": xreqid})
	} else {
		md = md.Copy()
		md.Set("x-request-id", xreqid)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
