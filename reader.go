// Package stringreader provides the Marshal struct to marshal data from a string-to-string hashmap.
package stringreader

import (
	"fmt"
	"reflect"
)

// Marshal can unmarshal data from a Source.
// See the UnmarshalContext function for details.
type Marshal struct {
	NameTag       string // Optional, tag to read name from
	StrictNameTag bool   // When false, allow fallback to field name

	ParserTag     string // tag to read parser from
	DefaultParser string // default parser to fall back to (optional)
	InlineParser  string // parser name to use for recursive struct parsing (optional)

	// Known set of parsers
	SingleParsers map[string]SingleParser
	MultiParsers  map[string]MultiParser

	// Use StrictTyping to prevent auto-conversion of returned values
	StrictTyping bool
}

// SingleParser is a function that parses a single value
type SingleParser = func(value string, ok bool, ctx ParsingContext) (interface{}, error)

// MultiParser is a function that parses multiple values
type MultiParser = func(value []string, ok bool, ctx ParsingContext) (interface{}, error)

// UnmarshalContext unmarshals data from source into dest.
//
// Dest must be a pointer to a struct; if this is not the case, ErrNotPointerToStruct is returned.
// Data is unmarshaled from source to dest as follows:
//
// For each public field, the go tags are examined.
//
// When m.InlineParser is non-empty and m.ParserTag is non-empty and the parser tag equals the inline tag, atttempt
// to recursivly calls UnmarshalContext with the same source and data.
// When the field type is a struct, the field value can be used as a new dest.
// When the field type is a pointer to a struct, create a new zero value (when needed) for the provided type and then use it as a dest.
// When the field type is none of the above, return ErrInlineNotStruct.
//
// When m.NameTag is non-empty, data from the specified name is read from source.
// When m.NameTag does not exist, and m.StrictNameTag is true, the field is skipped.
// When m.NameTag does not exist and m.StrictNameTag is false, data from the name of the field is read from source.
//
// When m.ParserTag is non-empty, the value and ok are passed to the defined function in m.SingleParsers or m.MultiParsers.
// When m.ParserTag is empty, and m.DefaultParser is non-empty, the value and ok are passed to the default function in m.SingleParsers or m.MultiParsers.
// When m.ParserTag is empty, and m.DefaultParser is empty, or the referenced parser function does not exist, an error is returned.
// When a parser exists in both m.SingleParsers and m.MultiParsers, an error is returned.
// When calling a parsing context, the ctx argument is passed to it unchanged.
//
// When the Parser function returns a value and nil error, it is written into the specified field of dest.
// When strict typing is disabled, will first attempt to convert the value to the target type.
// When either the conversion, or assignablity is impossible, an error is returned.
func (m Marshal) UnmarshalContext(dest interface{}, source Source, data ParsingData) error {
	// grab the pointer to the destination
	dPtr := reflect.ValueOf(dest)
	if dest == nil {
		return ErrValueIsNil
	}

	// check that the value is a pointer,
	// and the pointed to value is a struct.
	dType := dPtr.Type()
	if dType == nil || dType.Kind() != reflect.Ptr {
		return ErrNotPointerToStruct
	}
	dType = dType.Elem()
	if dType == nil || dType.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}

	// keep track of the destination value!
	dValue := dPtr.Elem()

	ctx := mutableParsingContext{data: data}

	hasInlineParser := m.InlineParser != ""

	// Read the fields of the type
	for i := 0; i < dType.NumField(); i++ {
		field := dType.Field(i)
		ctx.dest = field.Name

		// determine the type of validator to run
		// if we don't have a default type, don't do anything to fields without types!
		ctx.parser = field.Tag.Get(m.ParserTag)
		if ctx.parser == "" {
			if m.DefaultParser == "" {
				continue
			}
			ctx.parser = m.DefaultParser
		}

		// we have the inline parser value, so recursively process the struct
		// as instructed by the user.
		if hasInlineParser && ctx.parser == m.InlineParser {
			var fieldPointer interface{}

			// determine the type of the inlined field
			fType := field.Type
			switch fType.Kind() {

			// it is a struct (without an indirection) => simple
			case reflect.Struct:
				fieldPointer = dPtr.Elem().Field(i).Addr().Interface()

			// it should be a pointer to a struct
			case reflect.Ptr:
				// check that the pointed to element is indeed a struct
				fType = fType.Elem()
				if fType.Kind() != reflect.Struct {
					return ErrInlineNotStruct
				}

				// ensure that the value is not nil
				fValue := dPtr.Elem().Field(i)
				if fValue.IsNil() {
					fValue.Set(reflect.New(fType))
				}
				// and use the fieldPointer
				fieldPointer = fValue.Interface()
			default:
				return ErrInlineNotStruct
			}

			err := m.UnmarshalContext(fieldPointer, source, data)
			if err != nil {
				return err
			}
			continue
		}

		// determine the name to lookup in source
		// when we need a strict field, only use explicitly annotated ones
		ctx.source = field.Tag.Get(m.NameTag)
		if ctx.source == "" {
			if m.StrictNameTag {
				continue
			}
			ctx.source = field.Name
		}

		// figure out if we have a single or a multi parser
		singleParser, multiParser, err := m.GetParser(ctx.parser)
		if err != nil {
			return ErrUnknownParser{Parser: ctx.parser, Field: field.Name, cause: err}
		}

		// Get the appropriate field, and then parse it!
		var pValue interface{}
		var pErr error

		switch {
		case singleParser != nil:
			rValue, rOK := source.Get(ctx.source)
			pValue, pErr = singleParser(rValue, rOK, ctx)
		case multiParser != nil:
			rValue, rOK := source.GetAll(ctx.source)
			pValue, pErr = multiParser(rValue, rOK, ctx)
		}

		// trigger errors
		if pErr != nil {
			return ErrFailedToReadField{Field: ctx.dest, cause: pErr}
		}

		// convert the value that was returned to the appropriate type in the field!
		rValue := reflect.ValueOf(pValue)
		fValue := dValue.FieldByName(ctx.dest)
		fType := fValue.Type()

		// when we don't have strict typing, allow automatic converstion of the value
		if !m.StrictTyping {
			if rValue.IsValid() {
				if !rValue.CanConvert(fType) {
					return ErrNotConvertible{Field: field.Name, ReturnedType: rValue.Type(), FieldType: fType}
				}
				rValue, err = reflectConvert(rValue, fType)
				if err != nil {
					return ErrNotConvertible{Field: field.Name, ReturnedType: rValue.Type(), FieldType: fType, cause: err}
				}
			} else {
				// reflect.ValueOf(rValue) returned an invalid value.
				// this can only happen when rValue is the zero value.
				//
				// so magically assume the zero-value of the desired type instead.
				rValue = reflect.New(fType).Elem()
			}
		}

		// check if we can assign the value, then assign
		if !rValue.Type().AssignableTo(fType) {
			return ErrNotAssignable{Field: field.Name, ReturnedType: rValue.Type(), FieldType: fType}
		}
		fValue.Set(rValue)
	}
	return nil
}

// Unmarshal is like UnmarshalContext, but with a nil context
func (m Marshal) Unmarshal(dest interface{}, source Source) error {
	return m.UnmarshalContext(dest, source, ParsingData{})
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
func (m Marshal) UnmarshalSingle(dest interface{}, source SingleSource) error {
	return m.Unmarshal(dest, SourceFromSingle(source))
}

// UnmarshalMulti is like Unmarshal, except that it pretends the single fields for source do not exist
func (m Marshal) UnmarshalMulti(dest interface{}, source MultiSource) error {
	return m.Unmarshal(dest, SourceFromMulti(source))
}

// GetParser finds either a single or multi parser, and performs appropriate error checking
func (m Marshal) GetParser(name string) (single SingleParser, multi MultiParser, err error) {
	var singleOK, multiOK bool

	// find non-nil values in the parsers!
	single, singleOK = m.SingleParsers[name]
	multi, multiOK = m.MultiParsers[name]

	singleOK = singleOK && single != nil
	multiOK = multiOK && multi != nil

	// ensure that we have exactly one value, or fail
	if singleOK && multiOK {
		return nil, nil, ErrBothParserType
	}

	if !(singleOK || multiOK) {
		return nil, nil, ErrUnknownParserType
	}

	return
}

// RegisterSingleParser registers a new SingleParser with m.
//
// Parser should not be nil, and should not exist in m.MultiParsers.
// No checking of these conditions is performed; they should be ensured by the caller.
func (m *Marshal) RegisterSingleParser(name string, parser SingleParser) {
	if m.SingleParsers == nil {
		m.SingleParsers = make(map[string]SingleParser)
	}
	m.SingleParsers[name] = parser
}

// RegisterSingleParser registers a new MultiParser with m.
//
// Parser should not be nil, and should not exist in m.SingleParsers.
// No checking of these conditions is performed; they should be ensured by the caller.
func (m *Marshal) RegisterMultiParser(name string, parser MultiParser) {
	if m.MultiParsers == nil {
		m.MultiParsers = make(map[string]MultiParser)
	}
	m.MultiParsers[name] = parser
}
