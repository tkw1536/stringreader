package stringreader

// Source represents a source of string-identified data.
// Each datum is identified using a string key.
//
// Data exists in two forms, either a single string value or a slice of string values.
// These are defined in SourceSingle and SourceMulti represectively.
// Data between the two does not have to be related.
//
// To create a Source from SourceSingle and SourceMulti components, use SourceSplit.
type Source interface {
	SourceSingle
	SourceMulti
}

// SourceSingle represents a source of data with each datum being a single string.
// See also Source.
type SourceSingle interface {
	// Get attempts to read the datum with the provided key.
	// Returns the value of the datum and and the value true.
	//
	// When the provided datum does not exist,
	// or an error occurs attempting to read it, returns the empty string and false.
	Get(key string) (value string, ok bool)
}

// SourceMulti represents a source of data with each datum being a single string.
// See also Source.
type SourceMulti interface {
	// Get attempts to read the datum with the provided key.
	// Returns the value of the datum and and the value true.
	//
	// When the provided datum does not exist,
	// or an error occurs attempting to read it, returns nil and false.
	GetAll(key string) (value []string, ok bool)
}

// SourceSplit represents a Source that consists of a SourceSingle and a SourceMulti.
// Each source is used for their respective operations.
//
// When either ComponentSource is nil, simulates an empty source.
type SourceSplit struct {
	SourceSingle
	SourceMulti
}

func (s SourceSplit) Get(value string) (string, bool) {
	if s.SourceSingle == nil {
		return "", false
	}
	return s.SourceSingle.Get(value)
}

func (s SourceSplit) GetAll(value string) ([]string, bool) {
	if s.SourceMulti == nil {
		return nil, false
	}
	return s.SourceMulti.GetAll(value)
}

// SourceSingleMap implements SourceSingle.
type SourceSingleMap map[string]string

func (s SourceSingleMap) Get(src string) (value string, ok bool) {
	value, ok = s[src]
	return
}

// SourceMultiMap implements SourceMulti.
type SourceMultiMap map[string][]string

func (src SourceMultiMap) GetAll(key string) (value []string, ok bool) {
	value, ok = src[key]
	return
}
