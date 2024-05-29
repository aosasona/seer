package seer

import "strings"

var (
	collectRuntimeInfo = true
	defaultMessage     = "an error occurred"
)

type seer struct {
	op            string
	originalError error
}

/**
* WrapError is a function that takes an operation name, an original error, and a custom message and returns a new error with a more informative stack trace.
*
* Usage:
* ```go
* if _, err := doThing(); err != nil {
*   return seer.WrapError("doThing", err, "failed to do the thing")
* }
*````
**/
func WrapError(op string, originalError error, customMessage ..string) error {
	return nil
}

var Wrap = WrapError

// SetDefaultMessage sets the default message that will be used when a custom message is not provided.
func SetDefaultMessage(message string) {
	if strings.TrimSpace(message) != "" {
		defaultMessage = message
	}
}

// SetCollectRuntimeData controls collecting runtime info. Defaults to false
func SetCollectRuntimeData(flag bool) {
	collectRuntimeInfo = flag
}
