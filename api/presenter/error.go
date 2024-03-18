package presenter

import "github.com/heroiclabs/nakama-common/runtime"

var (
	ErrInternalError  = runtime.NewError("internal server error", 1) // INTERNAL
	ErrMarshal        = runtime.NewError("cannot marshal type", 2)   // INTERNAL
	ErrNoInputAllowed = runtime.NewError("no input allowed", 3)      // INVALID_ARGUMENT
	ErrNoUserIdFound  = runtime.NewError("no user ID in context", 4) // INVALID_ARGUMENT
	ErrUnmarshal      = runtime.NewError("cannot unmarshal type", 5) // INTERNAL
	ErrInvalidInput   = runtime.NewError("Invalid input", 6)
	ErrUserNotFound   = runtime.NewError("User not found", 6)

	ErrBetNotFound        = runtime.NewError("cannot find bet", 101) // INTERNAL
	ErrMatchNotFound      = runtime.NewError("cannot find match", 102)
	ErrNotEnoughChip      = runtime.NewError("not enough chip", 103)
	ErrFuncDisableByVipLv = runtime.NewError("function disable by vip lv", 104) // INTERNAL
	ErrNotFound           = runtime.NewError("not found", 105)

	ErrUserNameLenthTooShort       = runtime.NewError("Invalid username address, must be 8-255 bytes.", 1000)
	ErrUserNameLenthTooLong        = runtime.NewError("Invalid username address, must be 8-255 bytes.", 1001)
	ErrUserPasswordLenthTooShort   = runtime.NewError("Password must be at least 8 characters long.", 1002)
	ErrUserNameAndPasswordRequired = runtime.NewError("Username address and password is required", 1003)
	ErrUserNameExist               = runtime.NewError("Username already exist", 1004)
	ErrUserNameInvalid             = runtime.NewError("Invalid username, no spaces or control characters allowed.", 1005)
)
