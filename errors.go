package stringreader

import (
	"fmt"
	"reflect"

	"errors"
)

// MarshalError is an error that is returned during (un)marshaling
type MarshalError interface {
	error
	State
}

// freeUnmarshalError implements UnmarshalError, but does not contain any contextual information.
type freeUnmarshalError string

func (freeUnmarshalError) Field() string          { return "" }
func (freeUnmarshalError) Tag() reflect.StructTag { return "" }

func (freeUnmarshalError) Datum() string { return "" }

func (freeUnmarshalError) Typ() string { return "" }
func (freeUnmarshalError) Kind() Kind  { return KindUndef }

func (err freeUnmarshalError) Error() string { return string(err) }

var ErrDestIsNil MarshalError = freeUnmarshalError("Marshal: dest is nil")
var ErrNotPointerToStruct MarshalError = freeUnmarshalError("Marshal: value is not a pointer to a struct")

// ErrInlineNotStruct indicates that a destination field that is to be inlined, but is not a struct.
// Implements UnmarshalError.
type ErrInlineNotStruct struct {
	field string
	tag   reflect.StructTag

	typ string
}

func (err ErrInlineNotStruct) Field() string          { return err.field }
func (err ErrInlineNotStruct) Tag() reflect.StructTag { return err.tag }

func (ErrInlineNotStruct) Datum() string { return "" }

func (err ErrInlineNotStruct) Typ() string { return err.typ }
func (ErrInlineNotStruct) Kind() Kind      { return KindUndef }

func (err ErrInlineNotStruct) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Destination field %s is to be inlined, but not a struct or pointer to struct", err.field)
}

// ErrUnknownType indicates that an unknown typ was encountered.
// Implements MarshalError.
type ErrUnknownType struct {
	field string
	tag   reflect.StructTag

	datum string

	typ string

	cause error
}

func (err ErrUnknownType) Field() string          { return err.field }
func (err ErrUnknownType) Tag() reflect.StructTag { return err.tag }

func (err ErrUnknownType) Datum() string { return err.datum }

func (err ErrUnknownType) Typ() string { return err.typ }
func (ErrUnknownType) Kind() Kind      { return KindUndef }

func (err ErrUnknownType) Unwrap() error { return err.cause }
func (err ErrUnknownType) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Destination field %q has unknown type %s: %s", err.field, err.typ, err.cause.Error())
}

// ErrFailedToProcessField indicates that Marshal failed to process a field.
// Implements MarshalError.
type ErrFailedToProcessField struct {
	field string
	tag   reflect.StructTag

	datum string

	typ  string
	kind Kind

	cause error
}

func (err ErrFailedToProcessField) Field() string          { return err.field }
func (err ErrFailedToProcessField) Tag() reflect.StructTag { return err.tag }

func (err ErrFailedToProcessField) Datum() string { return err.datum }

func (err ErrFailedToProcessField) Typ() string { return err.typ }
func (err ErrFailedToProcessField) Kind() Kind  { return err.kind }

func (err ErrFailedToProcessField) Unwrap() error { return err.cause }
func (err ErrFailedToProcessField) Error() string {
	return fmt.Sprintf("Marshal.Unmarshal: Failed to process field %q: %s", err.field, err.cause.Error())
}

// ErrWrongDestType intends that the returned value can not be assigned or converted to the destination field.
// Implements UnmarshalError.
type ErrWrongDestType struct {
	field string
	tag   reflect.StructTag

	datum string

	typ  string
	kind Kind

	Assignment   bool // indicates if the failed operation was an assignment or converstion
	ReturnedType reflect.Type
	DestType     reflect.Type

	cause error
}

func (err ErrWrongDestType) Field() string          { return err.field }
func (err ErrWrongDestType) Tag() reflect.StructTag { return err.tag }

func (err ErrWrongDestType) Datum() string { return err.datum }

func (err ErrWrongDestType) Typ() string { return err.typ }
func (err ErrWrongDestType) Kind() Kind  { return err.kind }

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
	return fmt.Sprintf("Marshal: Failed to process value for field %q: Got type %s, but cannot %s to %s%s", err.field, err.ReturnedType, err.DestType, verb, suffix)
}

// the errors below never have any information associated with it.

var ErrUnknownTyp = errors.New("Marshal: unknown type")
var ErrBothTyp = errors.New("Marshal: type found in both Single and Multi")
