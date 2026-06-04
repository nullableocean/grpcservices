package mapping

import (
	"errors"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapErrorToGrpcStatusError(e error) error {
	if errors.Is(e, errs.ErrNotAllowed) {
		return status.Error(codes.PermissionDenied, e.Error())
	}

	if errors.Is(e, errs.ErrNotFound) {
		return status.Error(codes.NotFound, e.Error())
	}

	if errors.Is(e, errs.ErrIncorrectData) {
		return status.Error(codes.InvalidArgument, e.Error())
	}

	return status.Error(codes.Internal, e.Error())
}
