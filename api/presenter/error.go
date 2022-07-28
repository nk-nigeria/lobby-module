package presenter

import "github.com/heroiclabs/nakama-common/runtime"

var (
	ErrInternalError  = runtime.NewError("internal server error", 1) // INTERNAL
	ErrMarshal        = runtime.NewError("cannot marshal type", 2)   // INTERNAL
	ErrNoInputAllowed = runtime.NewError("no input allowed", 3)      // INVALID_ARGUMENT
	ErrNoUserIdFound  = runtime.NewError("no user ID in context", 4) // INVALID_ARGUMENT
	ErrUnmarshal      = runtime.NewError("cannot unmarshal type", 5) // INTERNAL

	ErrBetNotFound   = runtime.NewError("cannot find bet", 101) // INTERNAL
	ErrMatchNotFound = runtime.NewError("cannot find match", 102)
	ErrNotEnoughChip = runtime.NewError("not enough chip", 103) // INTERNAL

)
