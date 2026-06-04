package mapping

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsClientSideGrpcError проверяет, что ошибка является клиентской
func IsClientSideGrpcError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	code := st.Code()
	switch code {
	case
		codes.OK,
		codes.Canceled,
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.Unauthenticated:
		return true
	default:
		return false
	}
}
