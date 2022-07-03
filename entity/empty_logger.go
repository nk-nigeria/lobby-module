package entity

import (
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

type EmptyLogger struct{}

func (e *EmptyLogger) Debug(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/*
	Log a message with optional arguments at INFO level. Arguments are handled in the manner of fmt.Printf.
*/
func (e *EmptyLogger) Info(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/*
	Log a message with optional arguments at WARN level. Arguments are handled in the manner of fmt.Printf.
*/
func (e *EmptyLogger) Warn(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/*
	Log a message with optional arguments at ERROR level. Arguments are handled in the manner of fmt.Printf.
*/
func (e *EmptyLogger) Error(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/*
	Return a logger with the specified field set so that they are included in subsequent logging calls.
*/
func (e *EmptyLogger) WithField(key string, v interface{}) runtime.Logger {
	return e
}

/*
	Return a logger with the specified fields set so that they are included in subsequent logging calls.
*/
func (e *EmptyLogger) WithFields(fields map[string]interface{}) runtime.Logger {
	return e
}

/*
	Returns the fields set in this logger.
*/
func (e *EmptyLogger) Fields() map[string]interface{} {
	return nil
}
