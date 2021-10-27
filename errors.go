package stringreader

import (
	"fmt"
	"reflect"

	"errors"
)

// UnmarshalError is an error returned by the Unmarshal function.
type UnmarshalError interface {
	error
	UnmarshalContext
}

// freeUnmarshalError implements UnmarshalError, but does not contain any contextual information.
type freeUnmarshalError string

func (freeUnmarshalError) Dest() string   { return "" }
func (freeUnmarshalError) Source() string { return "" }
func (freeUnmarshalError) Parser() string { return "" }
func (freeUnmarshalError) Single() bool   { return false }

func (err freeUnmarshalError) Error() string { return string(err) }

var ErrDestIsNil UnmarshalError = freeUnmarshalError("Marshal.Unmarshal: dest is nil")
var ErrNotPointerToStruct UnmarshalError = freeUnmarshalError("Marshal.Unmarshal: dest is not a pointer to a struct")

// ErrInlineNotStruct indicates that a destination field that is to be inlined, but is not a struct.
// Implements UnmarshalError.
type ErrInlineNotStruct struct {
	dest   string
	parser string
}

func (err ErrInlineNotStruct) Dest() string   { return err.dest }
func (err ErrInlineNotStruct) Source() string { return "" }
func (err ErrInlineNotStruct) Parser() string { return "" }
func (err ErrInlineNotStruct) Single() bool   { return false }

func (err ErrInlineNotStruct) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Destination field %s is to be inlined, but not a struct or pointer to struct", err.dest)
}

// ErrUnknownParser indicates that an unknown parser was encountered.
// Implements UnmarshalError.
type ErrUnknownParser struct {
	dest, source, parser string // TODO: fixme
	cause                error
}

func (err ErrUnknownParser) Dest() string   { return err.dest }
func (err ErrUnknownParser) Source() string { return err.source }
func (err ErrUnknownParser) Parser() string { return err.parser }
func (err ErrUnknownParser) Single() bool   { return false }

func (err ErrUnknownParser) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Destination field %q has unknown parser %s: %s", err.dest, err.parser, err.cause.Error())
}

// Unwrap provides compatibility for Go 1.13 error chains
func (err ErrUnknownParser) Unwrap() error { return err.cause }

// ErrFailedToParseField indicates that Marshal.Unmarshal failed to read a field.
// Implements UnmarshalError.
type ErrFailedToParseField struct {
	dest, source, parser string
	single               bool

	cause error
}

func (err ErrFailedToParseField) Dest() string   { return err.dest }
func (err ErrFailedToParseField) Source() string { return err.source }
func (err ErrFailedToParseField) Parser() string { return err.parser }
func (err ErrFailedToParseField) Single() bool   { return err.single }

// Unwrap provides compatibility for Go 1.13 error chains.
func (err ErrFailedToParseField) Unwrap() error { return err.cause }

func (err ErrFailedToParseField) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Failed to parse field %q: %s", err.dest, err.cause.Error())
}

// ErrWrongDestType intends that the returned value can not be assigned or converted to the destination field.
// Implements UnmarshalError.
type ErrWrongDestType struct {
	dest, source, parser string
	single               bool

	Assignment   bool // indicates if the failed operation was an assignment or converstion
	ReturnedType reflect.Type
	DestType     reflect.Type

	cause error
}

func (err ErrWrongDestType) Dest() string   { return err.dest }
func (err ErrWrongDestType) Source() string { return err.source }
func (err ErrWrongDestType) Parser() string { return err.parser }
func (err ErrWrongDestType) Single() bool   { return err.single }

// Unwrap provides compatibility for Go 1.13 error chains.
func (err ErrWrongDestType) Unwrap() error { return err.cause }

func (err ErrWrongDestType) Error() string {
	var suffix string
	if err.cause != nil {
		suffix = ": " + err.cause.Error()
	}
	var verb string
	if err.Assignment {
		verb = "assign"
	} else {
		verb = "convert"
	}
	return fmt.Sprintf("Marshal.Unmarshal: Failed to process value for field %q: Parser returned type %s, but cannot %s to %s%s", err.dest, err.ReturnedType, err.DestType, verb, suffix)
}

// the errors below never have any information associated with it.

var ErrUnknownParserType = errors.New("Marshal.Unmarshal: unknown parser type")
var ErrBothParserType = errors.New("Marshal.Unmarshal: parser type in both Single and Multi")
