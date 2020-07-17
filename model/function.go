// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"sync"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v2/util"
)

// FunctionMap represents the collection of Function.
type FunctionMap struct {
	fns *sync.Map
}

// [string]govaluate.ExpressionFunction

// AddFunction adds an expression function.
func (fm *FunctionMap) AddFunction(name string, function govaluate.ExpressionFunction) {
	fm.fns.LoadOrStore(name, function)
}

// LoadFunctionMap loads an initial function map.
func LoadFunctionMap() FunctionMap {
	fm := &FunctionMap{}
	fm.fns = &sync.Map{}

	fm.AddFunction("keyMatch", util.KeyMatchFunc)
	fm.AddFunction("keyMatch2", util.KeyMatch2Func)
	fm.AddFunction("keyMatch3", util.KeyMatch3Func)
	fm.AddFunction("keyMatch4", util.KeyMatch4Func)
	fm.AddFunction("regexMatch", util.RegexMatchFunc)
	fm.AddFunction("ipMatch", util.IPMatchFunc)
	fm.AddFunction("globMatch", util.GlobMatchFunc)

	return *fm
}

// GetFunctions return a map with all the functions
func (fm *FunctionMap) GetFunctions() map[string]govaluate.ExpressionFunction {
	ret := make(map[string]govaluate.ExpressionFunction)

	fm.fns.Range(func(k interface{}, v interface{}) bool {
		ret[k.(string)] = v.(govaluate.ExpressionFunction)
		return true
	})

	return ret
}
