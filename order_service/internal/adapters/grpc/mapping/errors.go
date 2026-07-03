package mapping

import (
	"errors"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapGrpcStatusToError(e error) error {
	s, ok := status.FromError(e)
	if ok {
		switch s.Code() {
		case codes.NotFound:
			return fmt.Errorf("%w: %w", e, ports.ErrNotFound)
		case codes.PermissionDenied:
			return fmt.Errorf("%w: %w", e, errs.ErrNotAllowed)
		case codes.InvalidArgument:
			return fmt.Errorf("%w: %w", e, errs.ErrIncorrectData)
		}
	}

	return fmt.Errorf("%w", ports.ErrFailedClientRequest)
}

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

	if errors.Is(e, errs.ErrIdempotencyProcessing) {
		return status.Error(codes.Aborted, "failed idempotency")
	}

	return status.Error(codes.Internal, "server error")
}
