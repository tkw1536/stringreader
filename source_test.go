package stringreader

import "fmt"

func ExampleSourceSingleMap() {
	var source SourceSingle = SourceSingleMap(map[string]string{
		"key": "value",
	})

	// get an existing key
	keyValue, keyOK := source.Lookup("key")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "key", keyValue, keyOK)

	// get a non-existing key
	fakeValue, fakeOK := source.Lookup("fake")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "fake", fakeValue, fakeOK)

	// Output:
	// source.Lookup("key") value="value" ok=true
	// source.Lookup("fake") value="" ok=false
}

func ExampleSourceMulti() {
	// create a new source map
	var source SourceMulti = SourceMultiMap(map[string][]string{
		"key": {"another", "value"},
	})

	// get an existing key
	keyValue, keyOK := source.LookupAll("key")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "key", keyValue, keyOK)

	// get a non-existing key
	fakeValue, fakeOK := source.LookupAll("fake")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "fake", fakeValue, fakeOK)

	// Output:
	// source.LookupAll("key") value=[another value] ok=true
	// source.LookupAll("fake") value=[] ok=false
}

// Create a new SourceSplit consisting of a SourceSingle and SourceMulti.
func ExampleSourceSplit() {
	// Create a new SourceSplit, and set either component to an appropriate map.
	var source Source = SourceSplit{
		SourceSingle: SourceSingleMap(map[string]string{
			"key": "value",
		}),
		SourceMulti: SourceMultiMap(map[string][]string{
			"key": {"another", "value"},
		}),
	}

	sKeyValue, sKeyOK := source.Lookup("key")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "key", sKeyValue, sKeyOK)

	sFakeValue, sFakeOK := source.Lookup("fake")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "fake", sFakeValue, sFakeOK)

	mKeyValue, mKeyOK := source.LookupAll("key")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "key", mKeyValue, mKeyOK)

	mFakeValue, mFakeOK := source.LookupAll("fake")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "fake", mFakeValue, mFakeOK)

	// Output:
	// source.Lookup("key") value="value" ok=true
	// source.Lookup("fake") value="" ok=false
	// source.LookupAll("key") value=[another value] ok=true
	// source.LookupAll("fake") value=[] ok=false
}

// Creating an empty SourceSplit creates a source without any data inside of it.
func ExampleSourceSplit_empty() {
	// Create a new SourceSplit, but do not set either component.
	var source Source = SourceSplit{}

	// read keys from the SourceSingle and SourceMulti components.

	sKeyValue, sKeyOK := source.Lookup("key")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "key", sKeyValue, sKeyOK)

	sFakeValue, sFakeOK := source.Lookup("fake")
	fmt.Printf("source.Lookup(%q) value=%q ok=%t\n", "fake", sFakeValue, sFakeOK)

	mKeyValue, mKeyOK := source.LookupAll("key")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "key", mKeyValue, mKeyOK)

	mFakeValue, mFakeOK := source.LookupAll("fake")
	fmt.Printf("source.LookupAll(%q) value=%v ok=%t\n", "fake", mFakeValue, mFakeOK)

	// Output:
	// source.Lookup("key") value="" ok=false
	// source.Lookup("fake") value="" ok=false
	// source.LookupAll("key") value=[] ok=false
	// source.LookupAll("fake") value=[] ok=false
}

func ExampleSinkSingleMap() {
	var sink = SinkSingleMap(make(map[string]string))

	fmt.Println(sink.Set("hello", "world"))
	fmt.Printf("sink[%q] = %q\n", "hello", sink["hello"])

	// Output:
	// true
	// sink["hello"] = "world"
}

func ExampleSinkMultiMap() {
	var sink = SinkMultiMap(make(map[string][]string))

	fmt.Println(sink.SetAll("hello", []string{"world"}))
	fmt.Printf("sink[%q] = %q\n", "hello", sink["hello"])

	// Output:
	// true
	// sink["hello"] = ["world"]
}

// Create a new SinkSplit consisting of a SinkSingle and SinkMulti.
func ExampleSinkSplit() {
	single := SinkSingleMap(make(map[string]string))
	multi := SinkMultiMap(make(map[string][]string))

	// Create a new SourceSplit, and set either component to an appropriate map.
	var sink Sink = SinkSplit{
		SinkSingle: single,
		SinkMulti:  multi,
	}

	fmt.Println(sink.Set("hello", "world"))
	fmt.Printf("single[%q] = %q\n", "hello", single["hello"])

	fmt.Println(sink.SetAll("hello", []string{"world"}))
	fmt.Printf("multi[%q] = %q\n", "hello", multi["hello"])

	// Output:
	// true
	// single["hello"] = "world"
	// true
	// multi["hello"] = ["world"]
}

// Creating an empty SinkSplit creates a sink which always fails
func ExampleSinkSplit_empty() {
	// Create a new SinkSplit, but do not set either component.
	var sink Sink = SinkSplit{}

	fmt.Println(sink.Set("hello", "world"))
	fmt.Println(sink.SetAll("hello", []string{"world"}))

	// Output:
	// false
	// false
}
