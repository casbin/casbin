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

package log

import (
	"reflect"
	"testing"
)

type LoggerTester struct {
	event       int
	line        string
	lastMessage []interface{}
}

func (t *LoggerTester) EnableLog(bool)  {}
func (t *LoggerTester) IsEnabled() bool { return true }

func (t *LoggerTester) LogModel(event int, line []string, model [][]string) {
	t.event = event
	t.line = ""
	t.lastMessage = []interface{}{model}
}

func (t *LoggerTester) LogEnforce(event int, line string, request *[]interface{}, policies *[]string, result *[]interface{}) {
	t.event = event
	t.line = ""
	t.lastMessage = *request
}

func (t *LoggerTester) LogRole(event int, line string, role []string) {
	t.event = event
	t.line = ""
	t.lastMessage = []interface{}{role}
}

func (t *LoggerTester) LogPolicy(event int, line string, pPolicyFormat []string, gPolicyFormat []string, pPolicy *[]interface{}, gPolicy *[]interface{}) {
	t.event = event
	t.line = ""
	t.lastMessage = []interface{}{pPolicyFormat, gPolicyFormat, pPolicy, gPolicy}
}

func TestLog(t *testing.T) {
	lt := &LoggerTester{}
	SetLogger(lt)

	LogEnforce(1, "2", &[]interface{}{"3"}, &[]string{"4"}, &[]interface{}{true})
	if lt.event != 1 || lt.line != "" || !reflect.DeepEqual(lt.lastMessage, []interface{}{"3"}) {
		t.Errorf("incorrect logger message: %+v", lt.lastMessage)
	}

	LogModel(1, []string{"2"}, [][]string{{"3", "4"}})
	if lt.event != 1 || lt.line != "" || !reflect.DeepEqual(lt.lastMessage, []interface{}{[][]string{{"3", "4"}}}) {
		t.Errorf("incorrect logger message: %+v", lt.lastMessage)
	}
}
