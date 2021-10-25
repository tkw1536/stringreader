package stringreader

import "fmt"

func ExampleSourceSingleMap() {
	// create a new source map
	var source SourceSingle = SourceSingleMap(map[string]string{
		"key": "value",
	})

	// get an existing key
	keyValue, keyOK := source.Get("key")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "key", keyValue, keyOK)

	// get a non-existing key
	fakeValue, fakeOK := source.Get("fake")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "fake", fakeValue, fakeOK)

	// Output:
	// source.Get("key") value="value" ok=true
	// source.Get("fake") value="" ok=false
}

func ExampleSourceMulti() {
	// create a new source map
	var source SourceMulti = SourceMultiMap(map[string][]string{
		"key": {"another", "value"},
	})

	// get an existing key
	keyValue, keyOK := source.GetAll("key")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "key", keyValue, keyOK)

	// get a non-existing key
	fakeValue, fakeOK := source.GetAll("fake")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "fake", fakeValue, fakeOK)

	// Output:
	// source.GetAll("key") value=[another value] ok=true
	// source.GetAll("fake") value=[] ok=false
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

	sKeyValue, sKeyOK := source.Get("key")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "key", sKeyValue, sKeyOK)

	sFakeValue, sFakeOK := source.Get("fake")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "fake", sFakeValue, sFakeOK)

	mKeyValue, mKeyOK := source.GetAll("key")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "key", mKeyValue, mKeyOK)

	mFakeValue, mFakeOK := source.GetAll("fake")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "fake", mFakeValue, mFakeOK)

	// Output:
	// source.Get("key") value="value" ok=true
	// source.Get("fake") value="" ok=false
	// source.GetAll("key") value=[another value] ok=true
	// source.GetAll("fake") value=[] ok=false
}

// Creating an empty SourceSplit creates a source without any data inside of it.
func ExampleSourceSplit_empty() {
	// Create a new SourceSplit, but do not set either component.
	var source Source = SourceSplit{}

	// read keys from the SourceSingle and SourceMulti components.

	sKeyValue, sKeyOK := source.Get("key")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "key", sKeyValue, sKeyOK)

	sFakeValue, sFakeOK := source.Get("fake")
	fmt.Printf("source.Get(%q) value=%q ok=%t\n", "fake", sFakeValue, sFakeOK)

	mKeyValue, mKeyOK := source.GetAll("key")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "key", mKeyValue, mKeyOK)

	mFakeValue, mFakeOK := source.GetAll("fake")
	fmt.Printf("source.GetAll(%q) value=%v ok=%t\n", "fake", mFakeValue, mFakeOK)

	// Output:
	// source.Get("key") value="" ok=false
	// source.Get("fake") value="" ok=false
	// source.GetAll("key") value=[] ok=false
	// source.GetAll("fake") value=[] ok=false
}
