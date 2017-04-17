package casbin

// FunctionMap represents the functions that are used in the matchers.
type FunctionMap map[string]func(args ...interface{}) (interface{}, error)

func addFunction(fm FunctionMap, name string, function func(args ...interface{}) (interface{}, error)) {
	fm[name] = function
}

func loadFunctionMap() FunctionMap {
	fm := make(FunctionMap)

	addFunction(fm, "keyMatch", keyMatchFunc)

	return fm
}
