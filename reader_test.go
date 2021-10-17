package stringreader_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/tkw1536/stringreader"
)

func TestMarshal_UnmarshalSingle(t *testing.T) {

	// setupparser resets the parsers for m
	reset_parsers := func(m *stringreader.Marshal) {
		m.SingleParsers = nil
		m.MultiParsers = nil

		m.RegisterSingleParser("always", func(value string, ok bool) (interface{}, error) {
			return value, nil
		})
		m.RegisterSingleParser("default", func(value string, ok bool) (interface{}, error) {
			return "default", nil
		})
		m.RegisterSingleParser("special", func(value string, ok bool) (interface{}, error) {
			return fmt.Sprintf("special:%q", value), nil
		})
		m.RegisterSingleParser("never", func(value string, ok bool) (interface{}, error) {
			return nil, errors.New("never parser")
		})
	}

	tests := []struct {
		name     string
		marshal  stringreader.Marshal // ignores the existing fields
		src      map[string]string
		wantDest interface{}
		wantErr  bool
	}{
		{
			name: "complete test",
			marshal: stringreader.Marshal{
				NameTag:       "name",
				ParserTag:     "parser",
				DefaultParser: "always",
				StrictNameTag: false,
				StrictTyping:  false,
			},
			src: map[string]string{
				"NoTag":     "no-tag",
				"ParserTag": "parser-tag",
				"name":      "name-tag",
				"special":   "special value",
			},
			wantDest: struct {
				NoTag            string
				ParserTag        string `parser:"always"`
				NameTag          string `name:"name"`
				ParserAndNameTag string `parser:"special" name:"special"`
			}{
				NoTag:            "no-tag",
				ParserTag:        "parser-tag",
				NameTag:          "name-tag",
				ParserAndNameTag: "special:\"special value\"",
			},
		},
		{
			name: "no default parser skips untagged fields",
			marshal: stringreader.Marshal{
				NameTag:       "name",
				ParserTag:     "parser",
				StrictNameTag: false,
				StrictTyping:  false,
			},
			src: map[string]string{
				"ParserTag": "parser-tag-value",
				"name":      "name-tag-value",
				"special":   "special-value",
			},
			wantDest: struct {
				NoTag            string
				NameTag          string `name:"name"`
				ParserTag        string `parser:"always"`
				ParserAndNameTag string `parser:"special" name:"special"`
			}{
				ParserTag:        "parser-tag-value",
				ParserAndNameTag: "special:\"special-value\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.marshal
			reset_parsers(&m)

			// create a new element of the dest type!
			target := reflect.TypeOf(tt.wantDest)
			dest := reflect.New(target)

			err := m.UnmarshalSingle(dest.Interface(), stringreader.SourceMap(tt.src))
			if tt.wantErr {
				if err == nil {
					t.Error("wantErr = true, err = nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Marshal.UnmarshalSingle() err = %s, want = nil", err.Error())
			}

			gotDest := dest.Elem().Interface()
			if !reflect.DeepEqual(gotDest, tt.wantDest) {
				t.Errorf("Marshal.UnmarshalSingle() dest = %v, want = %v", gotDest, tt.wantDest)
			}
		})
	}
}

func ExampleMarshal_UnmarshalSingle() {

	// marshal is a simple marshal
	marshal := &stringreader.Marshal{
		NameTag: "read",

		ParserTag:     "type",
		DefaultParser: "string",
	}

	// UserProfile is an example struct to be unmarshaled below.
	// This uses the "read" and "type" tags defined above.
	type UserProfile struct {
		User     string `read:"user"`
		Hostname string `read:"host"`
		Port     uint16 `read:"port" type:"port"`
	}

	// define the types used for unmarshaling.

	// the "string" type accepts any string, and falls back to the empty string if it does not exist.
	// it is the default.
	marshal.RegisterSingleParser("string", func(value string, ok bool) (interface{}, error) {
		return value, nil
	})

	// the "port" type parses a port number.
	// It returns port 22
	marshal.RegisterSingleParser("port", func(value string, ok bool) (interface{}, error) {
		// if no port was provided, use port 22
		if !ok {
			return 22, nil
		}

		sport, err := strconv.ParseUint(value, 10, 16)
		return uint16(sport), err
	})

	// Parse a bunch of user profiles!

	var johnSmith UserProfile
	err := marshal.UnmarshalSingle(&johnSmith, stringreader.SourceMap(map[string]string{
		"port": "22",
		"user": "johnsmith",
		"host": "localhost",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &johnSmith)

	var janeSmith UserProfile
	err = marshal.UnmarshalSingle(&janeSmith, stringreader.SourceMap(map[string]string{
		"port": "2222",
		"user": "jane.smith",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &janeSmith)

	var jackSmith UserProfile
	err = marshal.UnmarshalSingle(&jackSmith, stringreader.SourceMap(map[string]string{
		"user": "jack.smith",
		"host": "localhost",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &jackSmith)

	// Output:
	// &{johnsmith localhost 22}
	// &{jane.smith  2222}
	// &{jack.smith localhost 22}
}
