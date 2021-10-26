package stringreader

import (
	"fmt"
	"reflect"

	"errors"
)

var ErrValueIsNil = errors.New("Marshal.Unmarshal: value is nil")
var ErrNotPointerToStruct = errors.New("Marshal.Unmarshal: dest is not a pointer to a struct")
var ErrInlineNotStruct = errors.New("Marshal.Unmarshal: inline field is not a struct or pointer to struct")
var ErrUnknownParserType = errors.New("Marshal.Unmarshal: unknown parser type")
var ErrBothParserType = errors.New("Marshal.Unmarshal: parser type in both Single and Multi")

// ErrUnknownParser indicates that Marshal.Unmarshal encountered an unknown parser
type ErrUnknownParser struct {
	Parser string
	Field  string
	cause  error
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (err ErrUnknownParser) Unwrap() error { return err.cause }

func (err ErrUnknownParser) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Type %q (for field %q) can not be determined: %s", err.Parser, err.Field, err.cause.Error())
}

// ErrFailedToReadField indicates that Marshal.Unmarshal failed to read a field
type ErrFailedToReadField struct {
	Field string
	cause error
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (err ErrFailedToReadField) Unwrap() error { return err.cause }

func (err ErrFailedToReadField) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Failed to read field %q: %s", err.Field, err.cause.Error())
}

// ErrNotConvertible intends that the returned value can not be assigned to the requested value.
type ErrNotConvertible struct {
	Field        string
	ReturnedType reflect.Type
	FieldType    reflect.Type
	cause        error
}

func (err ErrNotConvertible) Error() string {
	var suffix string
	if err.cause != nil {
		suffix = ": " + err.cause.Error()
	}
	return fmt.Sprintf("Marshal.Unmarshal: Failed to process value for field %q: Parser returned type %s, but cannot convert to %s%s", err.Field, err.ReturnedType, err.FieldType, suffix)
}

// ErrNotAssignable intends that the returned value can not be assigned to the requested value.
type ErrNotAssignable struct {
	Field        string
	ReturnedType reflect.Type
	FieldType    reflect.Type
}

func (err ErrNotAssignable) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Failed to process value for field %q: Parser returned type %s, but cannot assign to %s", err.Field, err.ReturnedType, err.FieldType)
}
