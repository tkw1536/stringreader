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

	// reset_unmarshalers resets the unmarshalers for m
	reset_unmarshalers := func(m *stringreader.Marshal) {
		m.SingleUnmarshalers = nil
		m.MultiUnmarshalers = nil

		m.RegisterSingleUnmarshaler("always", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
			return value, nil
		})
		m.RegisterSingleUnmarshaler("default", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
			return "default", nil
		})
		m.RegisterSingleUnmarshaler("special", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
			return fmt.Sprintf("special:%q", value), nil
		})
		m.RegisterSingleUnmarshaler("never", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
			return nil, errors.New("never unmarshaler")
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
				TypeTag:       "type",
				DefaultType:   "always",
				StrictNameTag: false,
				StrictTyping:  false,
			},
			src: map[string]string{
				"NoTag":   "no-tag",
				"TypeTag": "type-tag",
				"name":    "name-tag",
				"special": "special value",
			},
			wantDest: struct {
				NoTag          string
				TypeTag        string `type:"always"`
				NameTag        string `name:"name"`
				TypeAndNameTag string `type:"special" name:"special"`
			}{
				NoTag:          "no-tag",
				TypeTag:        "type-tag",
				NameTag:        "name-tag",
				TypeAndNameTag: "special:\"special value\"",
			},
		},
		{
			name: "no default type skips untagged fields",
			marshal: stringreader.Marshal{
				NameTag:       "name",
				TypeTag:       "type",
				StrictNameTag: false,
				StrictTyping:  false,
			},
			src: map[string]string{
				"TypeTag": "type-tag-value",
				"name":    "name-tag-value",
				"special": "special-value",
			},
			wantDest: struct {
				NoTag          string
				NameTag        string `name:"name"`
				TypeTag        string `type:"always"`
				TypeAndNameTag string `type:"special" name:"special"`
			}{
				TypeTag:        "type-tag-value",
				TypeAndNameTag: "special:\"special-value\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.marshal
			reset_unmarshalers(&m)

			// create a new element of the dest type!
			target := reflect.TypeOf(tt.wantDest)
			dest := reflect.New(target)

			err := m.UnmarshalSingle(dest.Interface(), stringreader.SourceSingleMap(tt.src))
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

func ExampleMarshal_UnmarshalSingle_simple() {

	// marshal is a simple marshal
	marshal := &stringreader.Marshal{
		NameTag: "read",

		TypeTag:     "type",
		DefaultType: "string",
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
	marshal.RegisterSingleUnmarshaler("string", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
		return value, nil
	})

	// the "port" type parses a port number.
	// It returns port 22
	marshal.RegisterSingleUnmarshaler("port", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
		// if no port was provided, use port 22
		if !ok {
			return 22, nil
		}

		sport, err := strconv.ParseUint(value, 10, 16)
		return uint16(sport), err
	})

	// Parse a bunch of user profiles!

	var johnSmith UserProfile
	err := marshal.UnmarshalSingle(&johnSmith, stringreader.SourceSingleMap(map[string]string{
		"port": "22",
		"user": "johnsmith",
		"host": "localhost",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &johnSmith)

	var janeSmith UserProfile
	err = marshal.UnmarshalSingle(&janeSmith, stringreader.SourceSingleMap(map[string]string{
		"port": "2222",
		"user": "jane.smith",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &janeSmith)

	var jackSmith UserProfile
	err = marshal.UnmarshalSingle(&jackSmith, stringreader.SourceSingleMap(map[string]string{
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

func ExampleMarshal_UnmarshalSingle_nil() {

	marshal := &stringreader.Marshal{
		NameTag: "read",

		TypeTag:     "type",
		DefaultType: "string",
	}

	type TheType struct {
		Value []string `type:"nilstringslice"`
	}
	marshal.RegisterSingleUnmarshaler("nilstringslice", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
		return nil, nil
	})

	var aType TheType
	err := marshal.UnmarshalSingle(&aType, stringreader.SourceSingleMap(map[string]string{}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", &aType)

	// Output:
	// &{[]}
}

func ExampleMarshal_UnmarshalSingle_recursive() {

	marshal := &stringreader.Marshal{
		NameTag: "read",

		TypeTag:    "type",
		InlineType: "inline",
	}

	// create three different nested structs.
	// that we can inline later

	type WithoutIndirection struct {
		Value string `read:"nested" type:"string"`
	}

	type WithIndirectionButNil struct {
		Value string `read:"pointed" type:"string"`
	}

	type WithIndirectionButNotNil struct {
		Value  string `read:"preset" type:"string"`
		Preset int
	}

	// create a new type, and register the string unmarshalers.
	// Inline the three nested structs
	type TheType struct {
		Plain         string                    `read:"plain" type:"string"`
		Inline        WithoutIndirection        `type:"inline"`
		NilPointer    *WithIndirectionButNil    `type:"inline"`
		NonNilPointer *WithIndirectionButNotNil `type:"inline"`
	}
	marshal.RegisterSingleUnmarshaler("string", func(value string, ok bool, ctx stringreader.Context) (interface{}, error) {
		return value, nil
	})

	// create a new element to read but preset the `Value` to 3.
	// this should prevent a new preset struct from being allocated.
	var aType TheType
	aType.NonNilPointer = &WithIndirectionButNotNil{
		Preset: 3,
	}
	err := marshal.UnmarshalSingle(&aType, stringreader.SourceSingleMap(map[string]string{
		"plain":   "plain value",
		"nested":  "inline value",
		"pointed": "pointed value",
		"preset":  "preset value",
	}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", aType.Plain)
	fmt.Printf("%v\n", aType.Inline)
	fmt.Printf("%v\n", *aType.NilPointer)
	fmt.Printf("%v\n", *aType.NonNilPointer)

	// Output:
	// plain value
	// {inline value}
	// {pointed value}
	// {preset value 3}
}
