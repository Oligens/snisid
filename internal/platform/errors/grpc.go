package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapToGRPCCode converts an internal ErrorCode to standard gRPC codes.
func MapToGRPCCode(code ErrorCode) codes.Code {
	switch code {
	case InvalidArgument:
		return codes.InvalidArgument
	case Unauthenticated:
		return codes.Unauthenticated
	case PermissionDenied:
		return codes.PermissionDenied
	case NotFound:
		return codes.NotFound
	case Conflict:
		return codes.AlreadyExists
	case Unavailable:
		return codes.Unavailable
	case Internal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// ToGRPCStatus securely translates a domain error into a gRPC status.
func ToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(*Error); ok {
		code := MapToGRPCCode(customErr.Code)

		if customErr.Code == Internal {
			// Mask message
			return status.Error(code, "An unexpected internal server error occurred")
		}

		return status.Error(code, customErr.Message)
	}

	// Default fallback for unknown errors
	return status.Error(codes.Unknown, "An unexpected error occurred")
}
