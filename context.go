package stringreader

// ParsingContext holds contextual data that is passed to parsers.
// It contains an internal reference to a ParsingData object.
//
// Parsers should not retain references to ParsingContext, as it might be re-used between parsers.
// See also SingleParser, MultiParser.
type ParsingContext interface {
	// Dest returns the name of the destination field that is being written to.
	Dest() string
	// Source returns the key of the datum that is being read.
	Source() string

	// Parser returns the name of the parser being used
	Parser() string
	// Single indicates if the parser being used is a SingleParser (true) or MultiParser (false)
	Single() bool

	// Get returns a local datum associated to the current destination from the underlying ParsingData object.
	Get(key string) interface{}

	// GetGlobal returns a global datum from the underlying ParsingData object.
	GetGlobal(key string) interface{}
}

// ParsingData holds contextual data for parsers.
// Data is keyed by strings.
// The zero value is ready-to-use.
//
// See also ParsingContext on how this data is accessed.
type ParsingData struct {
	// Globals holds data not associated to a specific field during parsing.
	// Each datum is keyed by a simple string.
	Globals map[string]interface{}

	// Locals holds data associated to a specific field and parser.
	// The key in the outer map corresponds to the name of the field.
	// The key in the inner map is like a key to Globals.
	Locals map[string]map[string]interface{}
}

// SetGlobal sets the global datum identified by key to value.
func (p *ParsingData) SetGlobal(key string, value interface{}) {
	if p.Globals == nil {
		p.Globals = make(map[string]interface{})
	}
	p.Globals[key] = value
}

// SetLocal sets the local datum for field identified by key to value.
func (p *ParsingData) SetLocal(field, key string, value interface{}) {
	if p.Locals == nil {
		p.Locals = make(map[string]map[string]interface{})
	}
	if _, ok := p.Locals[field]; !ok {
		p.Locals[field] = make(map[string]interface{})
	}
	p.Locals[field][key] = value
}

// parsingContext is the implementation of ParsingContext.
type parsingContext struct {
	dest, source, parser string
	single               bool
	data                 ParsingData
}

// Reset resets this parsing context to prepare it for re-use inside of a sync.Pool
func (p *parsingContext) Reset() {
	p.dest, p.source, p.parser = "", "", ""
	p.single = false
	p.data = ParsingData{}
}

// The remainder of functions implement ParsingContext.
// See the interface documentation for details.

func (p parsingContext) Dest() string {
	return p.dest
}

func (p parsingContext) Source() string {
	return p.source
}

func (p parsingContext) Parser() string {
	return p.parser
}

func (p parsingContext) Single() bool {
	return p.single
}

func (p parsingContext) GetGlobal(key string) interface{} {
	return p.data.Globals[key]
}

func (p parsingContext) Get(key string) interface{} {
	return p.data.Locals[p.dest][key]
}
