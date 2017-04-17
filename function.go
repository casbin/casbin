package casbin

// FunctionMap represents the collection of Function.
type FunctionMap map[string]func(args ...interface{}) (interface{}, error)

// Function represents a function that is used in the matchers, used to get attributes in ABAC.
type Function func(args ...interface{}) (interface{}, error)

func addFunction(fm FunctionMap, name string, function Function) {
	fm[name] = function
}

func loadFunctionMap() FunctionMap {
	fm := make(FunctionMap)

	addFunction(fm, "keyMatch", keyMatchFunc)

	return fm
}
