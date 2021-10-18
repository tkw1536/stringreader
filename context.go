package stringreader

// ParsingData holds data for parsers
type ParsingData struct {
	Globals map[string]interface{}
	Locals  map[string]map[string]interface{}
}

// SetGlobal sets a key in the global ParsingData
func (p *ParsingData) SetGlobal(key string, value interface{}) {
	if p.Globals == nil {
		p.Globals = make(map[string]interface{})
	}
	p.Globals[key] = value
}

// SetLocal sets a data local to a specific field
func (p *ParsingData) SetLocal(field, key string, value interface{}) {
	if p.Locals == nil {
		p.Locals = make(map[string]map[string]interface{})
	}
	if _, ok := p.Locals[field]; !ok {
		p.Locals[field] = make(map[string]interface{})
	}
	p.Locals[field][key] = value
}

// ParsingContext is passed to parsers
// Parsers may not retain it, and should copy any data neccessary.
type ParsingContext interface {
	// Dest returns the current field this parser is being used for.
	Dest() string

	// Source returns the name of the source field that is being read from.
	Source() string

	// Parser returns the name of the parser being used
	Parser() string

	// Get gets the contextual valiue for the current destination field and provided key
	Get(key string) interface{}

	// GetGlobal gets the contextual value for the provided global key
	GetGlobal(key string) interface{}
}

// implements ParsingContext
type mutableParsingContext struct {
	dest, source, parser string
	data                 ParsingData
}

func (p mutableParsingContext) Dest() string {
	return p.dest
}

func (p mutableParsingContext) Source() string {
	return p.source
}

func (p mutableParsingContext) Parser() string {
	return p.parser
}

// GetGlobal returns a global for the provided key
func (p mutableParsingContext) GetGlobal(key string) interface{} {
	return p.data.Globals[key]
}

// Get returns a value for the current destination field
func (p mutableParsingContext) Get(key string) interface{} {
	return p.data.Locals[p.dest][key]
}
