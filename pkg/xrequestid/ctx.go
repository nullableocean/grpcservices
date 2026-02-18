package xrequestid

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// извлекаем из контекста x-request-id
func GetXRequestIdFromCtx(ctx context.Context) string {
	meta, exist := metadata.FromIncomingContext(ctx)

	reqId := ""
	if exist {
		reqId = meta.Get(XREQUEST_ID_KEY)[0]
	}

	return reqId
}

// записываем x-request-id в контекст
func SetNewRequestIdToCtx(ctx context.Context) context.Context {
	xrequestId := NewXRequestId()
	md := metadata.New(map[string]string{"x-request-id": xrequestId})
	return metadata.NewOutgoingContext(ctx, md)
}
