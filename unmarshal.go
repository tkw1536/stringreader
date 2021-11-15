// Package stringreader provides the Marshal struct to marshal data from a string-to-string hashmap.
package stringreader

import (
	"fmt"
	"reflect"
	"sync"
)

// Marshal can marshal and unmarshal data from a string-to-string hashmap
// See the State function for details.
type Marshal struct {
	NameTag       string // Optional, tag to read name from
	StrictNameTag bool   // When false, allow fallback to field name

	TypeTag     string // tag to read type from
	DefaultType string // default type to fall back to (optional)
	InlineType  string // type name to use for recursive struct parsing (optional)

	// Known set of unmarshalers
	SingleUnmarshalers map[string]SingleUnmarshaler
	MultiUnmarshalers  map[string]MultiUnmarshaler

	// Use StrictTyping to prevent auto-conversion of returned values
	StrictTyping bool
}

// SingleUnmarshaler is a function that unmarshals a single value
type SingleUnmarshaler = func(value string, ok bool, ctx Context) (interface{}, error)

// MultiUnmarshaler is a function that parses multiple values
type MultiUnmarshaler = func(value []string, ok bool, ctx Context) (interface{}, error)

// a pool to receive Context objects from.
var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

// Unmarshal unmarshals data from source into value.
// Any non-nil error returned implements UnmarshalError.
//
// value must be a pointer to a struct; if this is not the case, ErrvalueIsNil or ErrNotPointerToStruct is returned.
//
// Data is unmarshaled from source to value as follows:
//
// For each public field, the go tags are examined.
//
// When m.InlineType is non-empty and m.TypeTag is non-empty and the type tag equals the inline tag, atttempt
// to recursivly calls Unmarshal with the same source and data.
// When the field type is a struct, the field value can be used as a new value.
// When the field type is a pointer to a struct, create a new zero value (when needed) for the provided type and then use it as a value.
// When the field type is none of the above, return ErrInlineNotStruct.
//
// When m.NameTag is non-empty, data from the specified name is read from source.
// When m.NameTag does not exist, and m.StrictNameTag is true, the field is skipped.
// When m.NameTag does not exist and m.StrictNameTag is false, data from the name of the field is read from source.
//
// When m.TypeTag is non-empty, the value and ok are passed to the defined function in m.SingleUnmarshalers or m.MultiUnmarshalers.
// When m.TypeTag is empty, and m.DefaultType is non-empty, the value and ok are passed to the default function in m.SingleUnmarshalers or m.MultiUnmarshalers.
// When m.TypeTag is empty, and m.DefaultType is empty, or the referenced unmarshaler function does not exist, an error is returned.
// When an unmarshaler exists in both m.SingleUnmarshalers and m.MultiUnmarshalers, an error is returned.
// When calling an unmarshaler, the data argument is passed to an appropriate context.
//
// When the Unmarshaler function returns a value and nil error, it is written into the specified field of value.
// When strict typing is disabled, will first attempt to convert the value to the target type.
// When either the conversion, or assignablity is impossible, an error is returned.
func (m Marshal) Unmarshal(value interface{}, source Source, data Data) error {
	return m.unmarshal(value, source, data)
}

// unmarshal implements Unmarshal, ensuring that an UnmarshalError is returned
func (m Marshal) unmarshal(value interface{}, source Source, data Data) MarshalError {
	if value == nil {
		return ErrDestIsNil
	}

	// ensure that the valueination is a pointer to a struct
	// and then use the pointer itself
	dValue := reflect.ValueOf(value)
	dType := dValue.Type()
	if dType == nil || dType.Kind() != reflect.Ptr {
		return ErrNotPointerToStruct
	}
	dType = dType.Elem()
	if dType == nil || dType.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}
	dValue = dValue.Elem()

	// grab a new context item from the pool
	// and store context data with it.
	ctx := contextPool.Get().(*context)
	defer contextPool.Put(ctx)

	ctx.data = data
	defer ctx.Reset()

	// Iterate over the values of that field
	dNum := dType.NumField()
	for i := 0; i < dNum; i++ {
		fStructField := dType.Field(i)
		fValue := dValue.Field(i)

		fType := fStructField.Type
		ctx.field = fStructField.Name
		ctx.tag = fStructField.Tag

		// determine the type to use
		// using the default type when necessary
		ctx.typ = fStructField.Tag.Get(m.TypeTag)
		if ctx.typ == "" {
			if m.DefaultType == "" {
				continue
			}
			ctx.typ = m.DefaultType
		}

		// check if the inline type is being requested.
		// and if so, do the inlining.
		if m.InlineType != "" && ctx.typ == m.InlineType {
			var fieldPointer interface{}

			switch fType.Kind() {
			// it is a struct (without an indirection) => simple
			case reflect.Struct:
				fieldPointer = fValue.Addr().Interface()

			// it should be a pointer to a struct
			case reflect.Ptr:
				// check that the pointed to element is indeed a struct
				fType = fType.Elem()
				if fType.Kind() != reflect.Struct {
					return ErrInlineNotStruct{
						field: ctx.field,
						tag:   ctx.tag,

						typ: ctx.typ,
					}
				}

				// when the value is nil, magically create a new value
				// so that we can fill zeroed pointer types.
				if fValue.IsNil() {
					fValue.Set(reflect.New(fType))
				}
				// and use the fieldPointer
				fieldPointer = fValue.Interface()
			default:
				return ErrInlineNotStruct{
					field: ctx.field,
					tag:   ctx.tag,

					typ: ctx.typ,
				}
			}

			err := m.unmarshal(fieldPointer, source, data)
			if err != nil {
				return err
			}
			continue
		}

		// determine which field to look at from the source
		// use default when needed
		ctx.datum = fStructField.Tag.Get(m.NameTag)
		if ctx.datum == "" {
			if m.StrictNameTag {
				continue
			}
			ctx.datum = fStructField.Name
		}

		// figure out if we have a single or a multi unmarshaler
		single, multi, err := m.getUnmarshaler(ctx.typ)
		if err != nil {
			return ErrUnknownType{
				field: ctx.field,
				tag:   ctx.tag,

				datum: ctx.datum,

				typ: ctx.typ,

				cause: err,
			}
		}

		// load and parse the appropriate value.
		var pValue interface{}
		var pErr error

		switch {
		case single != nil:
			rValue, rOK := source.Lookup(ctx.datum)
			ctx.kind = KindSingleUnmarshaler

			pValue, pErr = single(rValue, rOK, ctx)
		case multi != nil:
			rValue, rOK := source.LookupAll(ctx.datum)
			ctx.kind = KindMultiUnmarshaler

			pValue, pErr = multi(rValue, rOK, ctx)
		}
		if pErr != nil {
			return ErrFailedToProcessField{
				field: ctx.field,
				tag:   ctx.tag,

				datum: ctx.datum,

				typ:  ctx.typ,
				kind: ctx.kind,

				cause: pErr,
			}
		}

		// we need to convert the value we received to the proper type.
		rValue := reflect.ValueOf(pValue)

		if !m.StrictTyping {
			if rValue.IsValid() {
				// when we allow automatic type conversions and we have a valid (non-nil) value returned
				// convert the value to the proper type!
				if !rValue.CanConvert(fType) {
					return ErrWrongDestType{
						field: ctx.field,
						tag:   ctx.tag,

						datum: ctx.datum,

						typ:  ctx.typ,
						kind: ctx.kind,

						Assignment:   false,
						ReturnedType: rValue.Type(),
						DestType:     fType,

						cause: nil,
					}
				}
				rValue, err = reflectConvert(rValue, fType)
				if err != nil {
					return ErrWrongDestType{
						field: ctx.field,
						tag:   ctx.tag,

						datum: ctx.datum,

						typ:  ctx.typ,
						kind: ctx.kind,

						Assignment:   false,
						ReturnedType: rValue.Type(),
						DestType:     fType,

						cause: err,
					}
				}
			} else {
				// reflect.ValueOf(rValue) returned an invalid value.
				// this can only happen when rValue is the zero value.
				//
				// so magically assume the zero-value of the desired type instead.
				rValue = reflect.New(fType).Elem()
			}
		}

		// safely assign the value to the proper type!
		// we are already safe when we converterd
		if m.StrictTyping && !rValue.Type().AssignableTo(fType) {
			return ErrWrongDestType{
				field: ctx.field,
				tag:   ctx.tag,

				datum: ctx.datum,

				typ:  ctx.typ,
				kind: ctx.kind,

				Assignment:   true,
				ReturnedType: rValue.Type(),
				DestType:     fType,
			}
		}

		fValue.Set(rValue)
	}
	return nil
}

// UnmarshalAll is like State, but with a nil context
func (m Marshal) UnmarshalAll(value interface{}, source Source) error {
	return m.Unmarshal(value, source, Data{})
}

// reflectConvert converts rValue to rType, and catches any panic that occurs
func reflectConvert(rValue reflect.Value, rType reflect.Type) (v reflect.Value, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s", e)
		}
	}()

	return rValue.Convert(rType), nil
}

// UnmarshalSingle is like Unmarshal, except that it pretends the multi fields for source do not exist
func (m Marshal) UnmarshalSingle(value interface{}, source SourceSingle) error {
	return m.UnmarshalAll(value, SourceSplit{SourceSingle: source})
}

// UnmarshalMulti is like Unmarshal, except that it pretends the single fields for source do not exist
func (m Marshal) UnmarshalMulti(value interface{}, source SourceMulti) error {
	return m.UnmarshalAll(value, SourceSplit{SourceMulti: source})
}

// getUnmarshaler finds either a single or multi unmarshaler, and performs appropriate error checking
func (m Marshal) getUnmarshaler(name string) (single SingleUnmarshaler, multi MultiUnmarshaler, err error) {
	var singleOK, multiOK bool

	// find non-nil values in the unmarshalers!
	single, singleOK = m.SingleUnmarshalers[name]
	multi, multiOK = m.MultiUnmarshalers[name]

	singleOK = singleOK && single != nil
	multiOK = multiOK && multi != nil

	// ensure that we have exactly one value, or fail
	if singleOK && multiOK {
		return nil, nil, ErrBothTyp
	}

	if !(singleOK || multiOK) {
		return nil, nil, ErrUnknownTyp
	}

	return
}

// RegisterSingleUnmarshaler registers a new SingleUnmarshaler with m.
//
// unmarshaler should not be nil, and should not exist in m.MultiUnmarshalers.
// No checking of these conditions is performed; they should be ensured by the caller.
func (m *Marshal) RegisterSingleUnmarshaler(name string, unmarshaler SingleUnmarshaler) {
	if m.SingleUnmarshalers == nil {
		m.SingleUnmarshalers = make(map[string]SingleUnmarshaler)
	}
	m.SingleUnmarshalers[name] = unmarshaler
}

// RegisterSingleUnmarshaler registers a new MultiUnmarshaler with m.
//
// unmarshaler should not be nil, and should not exist in m.SingleUnmarshalers.
// No checking of these conditions is performed; they should be ensured by the caller.
func (m *Marshal) RegisterMultiUnmarshaler(name string, unmarshaler MultiUnmarshaler) {
	if m.MultiUnmarshalers == nil {
		m.MultiUnmarshalers = make(map[string]MultiUnmarshaler)
	}
	m.MultiUnmarshalers[name] = unmarshaler
}
