package stringreader

import "fmt"

func ExampleParsingData() {
	var data ParsingData

	// set a global and local key
	data.SetGlobal("world", 42)
	data.SetLocal("field", "world", 7)

	// read them out again
	fmt.Printf("data.Globals[%q] = %v\n", "world", data.Globals["world"])
	fmt.Printf("data.Locals[%q][%q] = %v\n", "field", "world", data.Locals["field"]["world"])

	// Output:
	// data.Globals["world"] = 42
	// data.Locals["field"]["world"] = 7
}
