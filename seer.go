package seer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
)

var (
	funcSuffixRegex   = regexp.MustCompile(`\.(func\d+)$`)
	collectStackTrace = true

	defaultCode    = 500
	defaultMessage = "an error occurred"
)

type Seer struct {
	op            string
	originalError error
	message       string
	code          int

	// runtime info
	caller string
	file   string
	line   int
}

type SeerInterface interface {
	error
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
}

// Code returns the error code that was set on the Seer error.
func (s *Seer) Code() int {
	if s.code == 0 {
		return defaultCode
	}

	return s.code
}

// WithCode sets the error code on the Seer error.
func (s *Seer) WithCode(code int) *Seer {
	if code < 100 || code > 599 {
		slog.Warn(fmt.Sprintf("Invalid error code %d. Not applying", code))
		return s
	}

	s.code = code
	return s
}

// Error returns our user-defined error message if it was set, otherwise it returns the original error message.
// WARNING: it returns the original error message for mostly use in logging, for user-facing error messages, use `Message` instead.
func (s *Seer) Error() string {
	if s.message == defaultMessage && s.originalError != nil {
		return s.originalError.Error()
	}

	return s.message // message is always set to either user-defined message or default message
}

// Message returns the user-defined error message that was passed to the Seer error, or the default message if a custom message was not provided.
func (s *Seer) Message() string {
	return s.message
}

// Operation returns the operation name that was passed to the Seer error.
func (s *Seer) Operation() string {
	return s.op
}

// OriginalError returns the original error that was wrapped by the Seer error.
func (s *Seer) OriginalError() error {
	return s.originalError
}

// ErrorWithStackTrace returns the error message with a bit more details, useful for debugging and logging
func (s *Seer) ErrorWithStackTrace() string {
	if collectStackTrace {
		callerName := s.caller

		var originalError string
		if s.originalError != nil {
			originalError = s.originalError.Error()
		}

		return fmt.Sprintf(
			"%s:%d (%s::%s): %s, original_err=%s",
			s.file,
			s.line,
			callerName,
			s.op,
			s.message,
			originalError,
		)
	} else {
		return fmt.Sprintf("%s: %s", s.op, s.message)
	}
}

// UnwrapError returns the original error and a boolean indicating whether the original error is a Seer error and can be further unwrapped.
func (s *Seer) UnwrapError() (error, bool) {
	_, nextErrorIsSeerError := s.originalError.(*Seer)
	return s.originalError, nextErrorIsSeerError
}

// String returns a string representation of the Seer error (stack trace included if `collectStackTrace` is set to true) that is useful for logging.
func (s *Seer) String() string {
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
func (s *Seer) MarshalJSON() ([]byte, error) {
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
func (s *Seer) UnmarshalJSON(data []byte) error {
	return nil
}

// New creates a new Seer error with the given operation name and message.
func New(operation string, message string, code ...int) *Seer {
	var (
		caller string
		file   string
		line   int
	)

	if collectStackTrace {
		caller, file, line = getRuntimeInfo()
	}

	errCode := 500
	if len(code) > 0 {
		errCode = code[0]

		if errCode < 100 || errCode > 599 {
			slog.
				Warn(fmt.Sprintf("Invalid error code %d. Defaulting to 500", errCode))
			errCode = 500
		}
	}

	return &Seer{
		op:      operation,
		message: message,
		code:    errCode,
		file:    file,
		caller:  caller,
		line:    line,
	}
}

/*
WrapError is a function that takes an operation name, an original error, and a custom message and returns a new error with a more informative stack trace.

Usage:

	if _, err := doThing(); err != nil {
	  return seer.Wrap("doThing", err, "failed to do the thing", 400)
	}
*/
func Wrap(op string, originalError error, customMessage ...string) *Seer {
	seerError := Seer{op: op, originalError: originalError}

	if len(customMessage) > 0 {
		seerError.message = customMessage[0]
	} else {
		seerError.message = defaultMessage
	}

	if collectStackTrace {
		seerError.caller, seerError.file, seerError.line = getRuntimeInfo()
	}

	return &seerError
}

// Unlike `Wrap`, `WrapWithStackTrace` is a function that takes an operation name, an original error, and a custom message and returns a new error with a more informative stack trace regardess of the `collectRuntimeInfo` flag.
// `operation` is the name of the operation that failed, to provide more context.
func WrapWithStackTrace(operation string, originalError error, customMessage ...string) *Seer {
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

	return &Seer{
		op:            operation,
		originalError: originalError,
		message:       message,
		file:          file,
		caller:        caller,
		line:          line,
	}
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

func SetDefaultCode(code int) {
	if code < 100 || code > 599 {
		slog.Warn(fmt.Sprintf("Invalid error code %d. Defaulting to 500", code))
		return
	}

	defaultCode = code
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

var (
	_ SeerInterface = (*Seer)(nil)
	_ error         = (*Seer)(nil)
)
