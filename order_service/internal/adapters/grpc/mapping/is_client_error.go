package mapping

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func IsGrpcClientErrors(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	code := st.Code()
	switch code {
	case codes.Canceled,
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
