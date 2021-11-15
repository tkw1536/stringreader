package stringreader

import "reflect"

// UnmarshalContext holds contextual data that is passed to parsers.
// It contains an internal reference to a UnmarshalerData object.
//
// Parsers should not retain references to UnmarshalContext, as it might be re-used between parsers.
// See also SingleParser, MultiParser.
type UnmarshalContext interface {
	UnmarshalState

	// Get returns a local datum associated to the current destination from the underlying UnmarshalerData object.
	Get(key string) interface{}

	// GetGlobal returns a global datum from the underlying UnmarshalerData object.
	GetGlobal(key string) interface{}
}

// UnmarshalState holds the current state of the unmarshaling process.
// It does not hold any references to UnmarshalerData objects.
type UnmarshalState interface {
	// Dest returns the name of the destination field that is being written to.
	// When no destination is being written, returns the empty string.
	Dest() string
	// Source returns the key of the datum that is being read.
	// When no destination is being written, returns the empty string.
	Source() string

	// Parser returns the name of the parser being used
	// When no parser is being used, returns the empty string.
	Parser() string
	// Single indicates if the parser being used is a SingleParser (true) or MultiParser (false).
	// When the Parser() method returns the empty string, the result is undefined.
	Single() bool

	// Tag returns the StructTag of the destination field that is being written to.
	Tag() reflect.StructTag
}

// UnmarshalerData holds contextual data for parsers.
// Data is keyed by strings.
// The zero value is ready-to-use.
//
// See also UnmarshalContext on how this data is accessed.
type UnmarshalerData struct {
	// Globals holds data not associated to a specific field during parsing.
	// Each datum is keyed by a simple string.
	Globals map[string]interface{}

	// Locals holds data associated to a specific field and parser.
	// The key in the outer map corresponds to the name of the field.
	// The key in the inner map is like a key to Globals.
	Locals map[string]map[string]interface{}
}

// SetGlobal sets the global datum identified by key to value.
func (p *UnmarshalerData) SetGlobal(key string, value interface{}) {
	if p.Globals == nil {
		p.Globals = make(map[string]interface{})
	}
	p.Globals[key] = value
}

// DeleteGlobal deletes the provided global datum identified by key.
func (p *UnmarshalerData) DeleteGlobal(key string) {
	if p.Globals == nil {
		return
	}
	delete(p.Globals, key)
}

// SetLocal sets the local datum for field identified by key to value.
func (p *UnmarshalerData) SetLocal(field, key string, value interface{}) {
	if p.Locals == nil {
		p.Locals = make(map[string]map[string]interface{})
	}
	if _, ok := p.Locals[field]; !ok {
		p.Locals[field] = make(map[string]interface{})
	}
	p.Locals[field][key] = value
}

// DeleteLocal deletes the provided local datum identified.
func (p *UnmarshalerData) DeleteLocal(field, key string) {
	locals, ok := p.Locals[field]
	if !ok || locals == nil {
		return
	}
	delete(locals, key)
}

// unmarshalContext is the implementation of UnmarshalContext.
type unmarshalContext struct {
	dest, source, parser string
	single               bool
	data                 UnmarshalerData
	tag                  reflect.StructTag
}

// Reset resets this parsing context to prepare it for re-use inside of a sync.Pool
func (p *unmarshalContext) Reset() {
	p.dest, p.source, p.parser = "", "", ""
	p.single = false
	p.data = UnmarshalerData{}
}

// The remainder of functions implement UnmarshalContext.
// See the interface documentation for details.

func (p unmarshalContext) Dest() string {
	return p.dest
}

func (p unmarshalContext) Source() string {
	return p.source
}

func (p unmarshalContext) Parser() string {
	return p.parser
}

func (p unmarshalContext) Single() bool {
	return p.single
}

func (p unmarshalContext) GetGlobal(key string) interface{} {
	return p.data.Globals[key]
}

func (p unmarshalContext) Get(key string) interface{} {
	return p.data.Locals[p.dest][key]
}

func (p unmarshalContext) Tag() reflect.StructTag {
	return p.tag
}
