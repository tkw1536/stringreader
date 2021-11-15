package stringreader

import "reflect"

// Context holds contextual data that is passed to (un)marshalers.
// It contains an internal reference to a Data object.
//
// (Un)Marshalers should not retain references to Context, as it might be re-used in between different processes.
// See also SingleUnmarshaler, MultiUnmarshaler.
type Context interface {
	State

	// Get returns the datum from the underlying Data associated with the current field being processed identified by the given key.
	// When no field is being processed, or no the provided value does not exist, returns nil.
	Get(key string) interface{}

	// GetGlobal returns the global datum from the underlying Data identified by the given key.
	// When no such datum exists, returns nil.
	GetGlobal(key string) interface{}
}

// State holds the current state of the (un)marshaling process.
// It does not hold any references to Data objects.
type State interface {
	// Field returns the name of the field that is being processed.
	// When no field is being processed, returns the empty string.
	Field() string
	// Tag returns the StructTag of the field that is being processed.
	Tag() reflect.StructTag

	// Datum returns the key of the datum that is being read.
	// When no destination is being written, returns the empty string.
	Datum() string

	// Typ returns the type of the field being processed.
	// When no field is being processed, returns the empty string.
	Typ() string
	// Kind indicates what kind of processor is being used
	Kind() Kind
}

// Kind represents which kind of (un)marshaling function is being used.
type Kind uint8

const (
	KindUndef Kind = iota

	KindSingleUnmarshaler
	KindMultiUnmarshaler

	KindSingleMarshaler
	KindMultiMarshaler
)

// Unmarshaler indicates if k is an unmarshaling kind
func (k Kind) Unmarshaler() bool {
	return k == KindSingleUnmarshaler || k == KindMultiUnmarshaler
}

func (k Kind) Marshaler() bool {
	return k == KindSingleMarshaler || k == KindMultiMarshaler
}

// Single indicates if k uses a single string as a value
func (k Kind) Single() bool {
	return k == KindSingleUnmarshaler || k == KindSingleMarshaler
}

// Multi indicates if k uses multiple strings as a value
func (k Kind) Multi() bool {
	return k == KindMultiUnmarshaler || k == KindMultiMarshaler
}

// Data holds contextual data passed to marshalers.
// Data is keyed by strings.
// The zero value is ready-to-use.
//
// See also Context on how this data is accessed.
type Data struct {
	// Globals holds data not associated to a specific field during parsing.
	// Each datum is keyed by a simple string.
	Globals map[string]interface{}

	// Locals holds data associated to a specific field and key.
	// The key in the outer map corresponds to the name typ.
	// The key in the inner map is like a key to Globals.
	Locals map[string]map[string]interface{}
}

// SetGlobal sets the global datum identified by key to value.
func (p *Data) SetGlobal(key string, value interface{}) {
	if p.Globals == nil {
		p.Globals = make(map[string]interface{})
	}
	p.Globals[key] = value
}

// SetLocal sets the local datum for field identified by key to value.
func (p *Data) SetLocal(field, key string, value interface{}) {
	if p.Locals == nil {
		p.Locals = make(map[string]map[string]interface{})
	}
	if _, ok := p.Locals[field]; !ok {
		p.Locals[field] = make(map[string]interface{})
	}
	p.Locals[field][key] = value
}

// context is the implementation of Context.
type context struct {
	field string
	tag   reflect.StructTag

	datum string

	typ  string
	kind Kind

	data Data
}

// Reset resets this parsing context to prepare it for re-use inside of a sync.Pool
func (p *context) Reset() {
	p.field, p.datum, p.typ = "", "", ""
	p.kind = KindUndef
	p.data = Data{}
}

// The remainder of functions implement Context.
// See the interface documentation for details.

func (p context) Field() string {
	return p.field
}

func (p context) Tag() reflect.StructTag {
	return p.tag
}

func (p context) Datum() string {
	return p.datum
}

func (p context) Typ() string {
	return p.typ
}

func (p context) Kind() Kind {
	return p.kind
}

func (p context) GetGlobal(key string) interface{} {
	return p.data.Globals[key]
}

func (p context) Get(key string) interface{} {
	return p.data.Locals[p.field][key]
}
