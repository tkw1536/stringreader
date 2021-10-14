package stringreader

// Source is a data source consisting of both a Single and MultiSource
type Source interface {
	SingleSource
	MultiSource
}

// SingleSource is a data source that takes as input a key and returns a single string value
type SingleSource interface {
	Get(key string) (value string, ok bool)
}

// MultiSource is a data source that takes as input a key and returns a single string value
type MultiSource interface {
	GetAll(key string) (value []string, ok bool)
}

// SourceFromSingle returns a new Source which uses a no-op multi source
func SourceFromSingle(s SingleSource) Source {
	return splitSource{single: s, multi: nopSource{}}
}

// SourceFromMulti returns a new Source which uses a no-op single source
func SourceFromMulti(m MultiSource) Source {
	return splitSource{single: nopSource{}, multi: m}
}

// splitSource represents a seperated single and multi source
type splitSource struct {
	single SingleSource
	multi  MultiSource
}

func (s splitSource) Get(value string) (string, bool) {
	return s.single.Get(value)
}

func (s splitSource) GetAll(value string) ([]string, bool) {
	return s.multi.GetAll(value)
}

// nopSource is a source that always returns false
type nopSource struct{}

func (nopSource) Get(value string) (string, bool) {
	return "", false
}

func (nopSource) GetAll(value string) ([]string, bool) {
	return nil, false
}

// SourceMap is a map that satfisfies Source
type SourceMap map[string]string

func (s SourceMap) Get(key string) (value string, ok bool) {
	value, ok = s[key]
	return
}
