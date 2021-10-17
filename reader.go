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

	// Known set of parsers
	SingleParsers map[string]SingleParser
	MultiParsers  map[string]MultiParser

	// Use StrictTyping to prevent auto-conversion of returned values
	StrictTyping bool
}

// ParsingContext is anything that can be used as a context for parsing
type ParsingContext func(key string) interface{}

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
func (m Marshal) UnmarshalContext(dest interface{}, source Source, ctx ParsingContext) error {
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

	// Read the fields of the type
	for i := 0; i < dType.NumField(); i++ {
		field := dType.Field(i)

		// determine the name to lookup in source
		// when we need a strict field, only use explicitly annotated ones
		fieldName := field.Tag.Get(m.NameTag)
		if fieldName == "" {
			if m.StrictNameTag {
				continue
			}
			fieldName = field.Name
		}

		// determine the type of validator to run
		// if we don't have a default type, don't do anything to fields without types!
		fieldKind := field.Tag.Get(m.ParserTag)
		if fieldKind == "" {
			if m.DefaultParser == "" {
				continue
			}
			fieldKind = m.DefaultParser
		}

		// figure out if we have a single or a multi parser
		singleParser, multiParser, err := m.GetParser(fieldKind)
		if err != nil {
			return ErrUnknownParser{Type: fieldKind, Field: field.Name, cause: err}
		}

		// Get the appropriate field, and then parse it!
		var pValue interface{}
		var pErr error

		switch {
		case singleParser != nil:
			rValue, rOK := source.Get(fieldName)
			pValue, pErr = singleParser(rValue, rOK, ctx)
		case multiParser != nil:
			rValue, rOK := source.GetAll(fieldName)
			pValue, pErr = multiParser(rValue, rOK, ctx)
		}

		// trigger errors
		if pErr != nil {
			return ErrFailedToReadField{Field: field.Name, cause: pErr}
		}

		// convert the value that was returned to the appropriate type in the field!
		rValue := reflect.ValueOf(pValue)
		fValue := dValue.FieldByName(field.Name)
		fType := fValue.Type()

		// when we don't have strict typing, allow automatic converstion of the value
		if !m.StrictTyping {
			if !rValue.CanConvert(fType) {
				return ErrNotConvertible{Field: field.Name, ReturnedType: rValue.Type(), FieldType: fType}
			}
			rValue, err = reflectConvert(rValue, fType)
			if err != nil {
				return ErrNotConvertible{Field: field.Name, ReturnedType: rValue.Type(), FieldType: fType, cause: err}
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
	return m.UnmarshalContext(dest, source, nil)
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
