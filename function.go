package casbin

import "github.com/hsluoyz/casbin/util"

// FunctionMap represents the collection of Function.
type FunctionMap map[string]func(args ...interface{}) (interface{}, error)

// Function represents a function that is used in the matchers, used to get attributes in ABAC.
type Function func(args ...interface{}) (interface{}, error)

// Add an expression function.
func (fm FunctionMap) AddFunction(name string, function Function) {
	fm[name] = function
}

// Load an initial function map.
func LoadFunctionMap() FunctionMap {
	fm := make(FunctionMap)

	fm.AddFunction("keyMatch", util.KeyMatchFunc)
	fm.AddFunction("regexMatch", util.RegexMatchFunc)

	return fm
}
