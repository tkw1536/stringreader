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

// SourceMulti represents a source of data with each datum being a single string.
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
// When either ComponentSource is nil, simulates an empty source.
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

// SourceSingleMap implements SourceSingle.
type SourceSingleMap map[string]string

func (s SourceSingleMap) Lookup(src string) (value string, ok bool) {
	value, ok = s[src]
	return
}

// SourceMultiMap implements SourceMulti.
type SourceMultiMap map[string][]string

func (src SourceMultiMap) LookupAll(key string) (value []string, ok bool) {
	value, ok = src[key]
	return
}

// SourceSmartSplit is like SourceSplit, but differs in behavior for unset components.
//
// When a component is unset, attempts to use the other component.
//
//  - a SingleSource is emulated using the first available element from the MultiSource
//  - a MultiSource is emulated returning either only the SingleSource element or nothing
//
// When neither component is present, returns an empty source.
type SourceSmartSplit struct {
	SourceSingle
	SourceMulti
}

func (s SourceSmartSplit) Lookup(value string) (string, bool) {
	switch {
	case s.SourceSingle != nil:
		return s.SourceSingle.Lookup(value)
	case s.SourceMulti != nil:
		result, ok := s.SourceMulti.LookupAll(value)
		if !ok || len(result) == 0 {
			return "", false
		}
		return result[0], true
	default:
		return "", false
	}
}

func (s SourceSmartSplit) LookupAll(value string) ([]string, bool) {
	switch {
	case s.SourceMulti != nil:
		return s.SourceMulti.LookupAll(value)
	case s.SourceSingle != nil:
		result, ok := s.SourceSingle.Lookup(value)
		if !ok {
			return nil, true
		}
		return []string{result}, true
	default:
		return nil, false
	}
}
