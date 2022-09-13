package api

import (
	"runtime/debug"

	"github.com/heroiclabs/nakama-common/runtime"
)

func Recovery(logger runtime.Logger) {
	if r := recover(); r != nil {
		logger.Info("Recovered. Error: %v\n Stack %s", string(debug.Stack()))
	}
}
