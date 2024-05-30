package seer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

var (
	funcSuffixRegex   = regexp.MustCompile(`\.(func\d+)$`)
	collectStackTrace = true
	defaultMessage    = "an error occurred"
)

type Seer struct {
	op            string
	originalError error
	message       string

	// runtime info
	caller string
	file   string
	line   int
}

type SeerInterface interface {
	error
	fmt.Stringer
	json.Marshaler
}

// Error returns our user-defined error message, useful for direct responses to the user.
func (s Seer) Error() string {
	return s.message
}

// ErrorWithStackTrace returns the error message with a bit more details, useful for debugging and logging
func (s Seer) ErrorWithStackTrace() string {
	if collectStackTrace {
		callerName := s.caller
		return fmt.Sprintf("%s:%d (%s::%s): %s", s.file, s.line, callerName, s.op, s.message)
	} else {
		return fmt.Sprintf("%s: %s", s.op, s.message)
	}
}

// UnwrapError returns the original error and a boolean indicating whether the original error is a Seer error and can be further unwrapped.
func (s Seer) UnwrapError() (error, bool) {
	_, nextErrorIsSeerError := s.originalError.(Seer)
	return s.originalError, nextErrorIsSeerError
}

func (s Seer) String() string {
	var sb strings.Builder

	defer sb.Reset() // Deallocate the string builder

	if collectStackTrace {
		callerName := s.caller
		sb.WriteString(fmt.Sprintf("%s:%d (%s::%s)", s.file, s.line, callerName, s.op))
	} else {
		sb.WriteString(fmt.Sprintf("%s: %s", s.op, s.message))
	}

	if s.originalError != nil {
		sb.WriteString(fmt.Sprintf("\n\tWrapped error: %s", s.originalError.Error()))
	}

	return sb.String()
}

// MarshalJSON returns a JSON representation of the Seer error, satisfying the json.Marshaler interface.
func (s Seer) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["operation"] = s.op
	data["message"] = s.message

	if collectStackTrace && (s.caller != "" || s.file != "" || s.line != 0) {
		data["caller"] = s.caller
		data["file"] = s.file
		data["line"] = s.line
	}

	if s.originalError != nil {
		data["previous_error"] = s.originalError.Error()
	}

	return json.Marshal(data)
}

// UnmarshalJSON is a no-op function that satisfies the json.Unmarshaler interface.
func (s Seer) UnmarshalJSON(data []byte) error {
	return nil
}

// New creates a new Seer error with the given operation name and message.
func New(op string, message string) error {
	var (
		caller string
		file   string
		line   int
	)

	if collectStackTrace {
		caller, file, line = getRuntimeInfo()
	}

	return Seer{op: op, message: message, file: file, caller: caller, line: line}
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
func Wrap(op string, originalError error, customMessage ...string) error {
	seerError := Seer{op: op, originalError: originalError}

	if len(customMessage) > 0 {
		seerError.message = customMessage[0]
	} else {
		seerError.message = defaultMessage
	}

	if collectStackTrace {
		seerError.caller, seerError.file, seerError.line = getRuntimeInfo()
	}

	return seerError
}

// Unlike `QuickWrap`, `WrapWithStackTrace` is a function that takes an operation name, an original error, and a custom message and returns a new error with a more informative stack trace regardess of the `collectRuntimeInfo` flag.
func WrapWithStackTrace(op string, originalError error, customMessage ...string) error {
	var (
		caller  string
		file    string
		line    int
		message string
	)

	if len(customMessage) > 0 {
		message = customMessage[0]
	} else {
		message = defaultMessage
	}

	caller, file, line = getRuntimeInfo()

	return Seer{op: op, originalError: originalError, message: message, file: file, caller: caller, line: line}
}

// SetDefaultMessage sets the default message that will be used when a custom message is not provided.
func SetDefaultMessage(message string) {
	if strings.TrimSpace(message) != "" {
		defaultMessage = message
	}
}

// SetCollectRuntimeData controls collecting runtime info. Defaults to false
func SetCollectStackTrace(flag bool) {
	collectStackTrace = flag
}

func getRuntimeInfo() (string, string, int) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "", "", 0
	}

	caller := runtime.FuncForPC(pc).Name()

	// Extract the function name from the caller (thing.Foo.func1 -> Foo)
	caller = caller[strings.LastIndex(caller, "/")+1:]

	// Remove the funcN suffix
	caller = funcSuffixRegex.ReplaceAllString(caller, "")

	return caller, file, line
}

var _ SeerInterface = Seer{}
