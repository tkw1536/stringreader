package stringreader_test

import (
	"fmt"
	"strconv"

	"github.com/tkw1536/stringreader"
)

func ExampleMarshal_UnmarshalSingle() {

	// marshal is a simple marshal
	marshal := &stringreader.Marshal{
		NameTag: "read",

		ParserTag:     "type",
		DefaultParser: "string",

		StrictTyping: false,
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
