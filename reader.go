// Package stringreader provides the Marshal struct to marshal data from a string-to-string hashmap.
package stringreader

import (
	"fmt"
	"reflect"
	"sync"
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

// a pool to receive ParsingContext objects from.
var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(parsingContext)
	},
}

// UnmarshalContext unmarshals data from source into dest.
// Any non-nil error returned implements UnmarshalError.
//
// Dest must be a pointer to a struct; if this is not the case, ErrDestIsNil or ErrNotPointerToStruct is returned.
//
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
	if dest == nil {
		return ErrDestIsNil
	}

	// ensure that the destination is a pointer to a struct
	// and then use the pointer itself
	dValue := reflect.ValueOf(dest)
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
	ctx := contextPool.Get().(*parsingContext)
	defer contextPool.Put(ctx)

	ctx.data = data
	defer ctx.Reset()

	dNum := dType.NumField()
	hasInlineParser := m.InlineParser != ""

	// Read the fields of the type
	for i := 0; i < dNum; i++ {
		fStructField := dType.Field(i)
		fType := fStructField.Type
		fValue := dValue.Field(i)

		ctx.dest = fStructField.Name

		// determine the type of parser to run
		// using the default type when necessary
		ctx.parser = fStructField.Tag.Get(m.ParserTag)
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
						dest:   ctx.dest,
						parser: ctx.parser,
					}
				}

				// ensure that the value is not nil
				if fValue.IsNil() {
					fValue.Set(reflect.New(fType))
				}
				// and use the fieldPointer
				fieldPointer = fValue.Interface()
			default:
				return ErrInlineNotStruct{
					dest:   ctx.dest,
					parser: ctx.parser,
				}
			}

			err := m.UnmarshalContext(fieldPointer, source, data)
			if err != nil {
				return err
			}
			continue
		}

		// determine the name to lookup in source
		// when we need a strict field, only use explicitly annotated ones
		ctx.source = fStructField.Tag.Get(m.NameTag)
		if ctx.source == "" {
			if m.StrictNameTag {
				continue
			}
			ctx.source = fStructField.Name
		}

		// figure out if we have a single or a multi parser
		singleParser, multiParser, err := m.GetParser(ctx.parser)
		if err != nil {
			return ErrUnknownParser{
				dest:   ctx.dest,
				source: ctx.source,
				parser: ctx.parser,

				cause: err,
			}
		}

		// Get the appropriate field, and then parse it!
		var pValue interface{}
		var pErr error

		switch {
		case singleParser != nil:
			rValue, rOK := source.Get(ctx.source)
			ctx.single = true

			pValue, pErr = singleParser(rValue, rOK, ctx)
		case multiParser != nil:
			rValue, rOK := source.GetAll(ctx.source)
			ctx.single = false

			pValue, pErr = multiParser(rValue, rOK, ctx)
		}
		if pErr != nil {
			return ErrFailedToParseField{
				dest:   ctx.dest,
				source: ctx.source,
				parser: ctx.parser,
				single: ctx.single,

				cause: pErr,
			}
		}

		// convert the value that was returned to the appropriate type in the field!
		rValue := reflect.ValueOf(pValue)

		if !m.StrictTyping {
			// when we allow automatic type conversions and we have a valid (non-nil) value returned
			// convert the value to the proper type!
			if rValue.IsValid() {
				if !rValue.CanConvert(fType) {
					return ErrWrongDestType{
						dest:   ctx.dest,
						source: ctx.source,
						parser: ctx.parser,
						single: ctx.single,

						Assignment:   false,
						ReturnedType: rValue.Type(),
						DestType:     fType,

						cause: nil,
					}
				}
				rValue, err = reflectConvert(rValue, fType)
				if err != nil {
					return ErrWrongDestType{
						dest:   ctx.dest,
						source: ctx.source,
						parser: ctx.parser,
						single: ctx.single,

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
		if !rValue.Type().AssignableTo(fType) {
			return ErrWrongDestType{
				dest:   ctx.dest,
				source: ctx.source,
				parser: ctx.parser,
				single: ctx.single,

				Assignment:   true,
				ReturnedType: rValue.Type(),
				DestType:     fType,
			}
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
func (m Marshal) UnmarshalSingle(dest interface{}, source SourceSingle) error {
	return m.Unmarshal(dest, SourceSplit{SourceSingle: source})
}

// UnmarshalMulti is like Unmarshal, except that it pretends the single fields for source do not exist
func (m Marshal) UnmarshalMulti(dest interface{}, source SourceMulti) error {
	return m.Unmarshal(dest, SourceSplit{SourceMulti: source})
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
