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
	// Lookup attempts to read the datum with the provided key.
	// Returns the value of the datum and and the value true.
	//
	// When the provided datum does not exist,
	// or an error occurs attempting to read it, returns the empty string and false.
	Lookup(key string) (value string, ok bool)
}

// SourceMulti represents a source of data with each datum being a string slice.
// See also Source.
type SourceMulti interface {
	// LookupAll attempts to read the datum with the provided key.
	// Returns the value of the datum and and the boolean true.
	//
	// When the provided datum does not exist,
	// or an error occurs attempting to read it, returns nil and false.
	LookupAll(key string) (value []string, ok bool)
}

// SourceSplit represents a Source that consists of a SourceSingle and a SourceMulti.
// Each source is used for their respective operations.
//
// When either component is nil, simulates an empty source.
type SourceSplit struct {
	SourceSingle
	SourceMulti
}

func (s SourceSplit) Lookup(value string) (string, bool) {
	if s.SourceSingle == nil {
		return "", false
	}
	return s.SourceSingle.Lookup(value)
}

func (s SourceSplit) LookupAll(value string) ([]string, bool) {
	if s.SourceMulti == nil {
		return nil, false
	}
	return s.SourceMulti.LookupAll(value)
}

// SourceSingleMap implements SourceSingle
type SourceSingleMap map[string]string

var _ SourceSingle = SourceSingleMap(nil)

func (s SourceSingleMap) Lookup(key string) (value string, ok bool) {
	value, ok = s[key]
	return
}

// SourceMultiMap implements SourceMulti.
type SourceMultiMap map[string][]string

var _ SourceMulti = SourceMultiMap(nil)

func (s SourceMultiMap) LookupAll(key string) (value []string, ok bool) {
	value, ok = s[key]
	return
}

// Sink represents a sink of string-identified data.
// Each datum is identified using a string key.
//
// Data exists in two forms, either a single string value or a slice of string values.
// These are defined in SinkSingle and SinkMulti represectively.
// Data between the two does not have to be related.
//
// To create a Sink from SinkSingle and SinkMulti components, use SinkSplit.
type Sink interface {
	SinkSingle
	SinkMulti
}

// SinkSingle represents a sink of data with each datum being a single string.
// See also Sink.
type SinkSingle interface {
	// Set sets the datum identified by key to be value.
	// Returns true if storing suceeded, and false when storing failed.
	Set(key string, value string) (ok bool)
}

// SinkMulti represents a sink of data with each datum being string slice.
// See also Sink.
type SinkMulti interface {
	// SetAll sets the datum identified by key to be value.
	// Returns true if storing succeeded, and false when storing failed.
	SetAll(key string, value []string) (ok bool)
}

// SinkSplit represents a Sink that consists of a SinkSingle and a SinkMulti.
// Each sink is used for their respective operations.
//
// When either component is nil, indicates failure.
type SinkSplit struct {
	SinkSingle
	SinkMulti
}

func (s SinkSplit) Set(key, value string) bool {
	if s.SinkSingle == nil {
		return false
	}
	return s.SinkSingle.Set(key, value)
}

func (s SinkSplit) SetAll(key string, value []string) bool {
	if s.SinkMulti == nil {
		return false
	}
	return s.SinkMulti.SetAll(key, value)
}

// SinkSingleMap implements SinkSingle
type SinkSingleMap map[string]string

var _ SinkSingle = SinkSingleMap(nil)

func (s SinkSingleMap) Set(key, value string) bool {
	s[key] = value
	return true
}

// SinkMultiMap implements SinkMulti.
type SinkMultiMap map[string][]string

var _ SinkMulti = SinkMultiMap(nil)

func (s SinkMultiMap) SetAll(key string, value []string) bool {
	s[key] = value
	return true
}
