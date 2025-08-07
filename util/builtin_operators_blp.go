// Copyright 2025 The casbin Authors. All Rights Reserved.
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

package util

import (
	"fmt"
)

// levelMatch enforces Bell-LaPadula security model.
// Simple Security Property: no read-up.
// Star Property: no write-down.
func levelMatch(action string, subjectLevel, objectLevel interface{}) bool {
	subLevel := int(subjectLevel.(float64))
	objLevel := int(objectLevel.(float64))

	if action == "read" {
		return subLevel >= objLevel
	}
	if action == "write" {
		return subLevel <= objLevel
	}
	return false
}

// LevelMatchFunc is the wrapper for levelMatch.
func LevelMatchFunc(args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return false, fmt.Errorf("levelMatch: expected 3 arguments (action, subjectLevel, objectLevel), but got %d", len(args))
	}

	action, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("levelMatch: action argument must be a string")
	}

	subjectLevel, ok := args[1].(float64)
	if !ok {
		return false, fmt.Errorf("levelMatch: subjectLevel argument must be a number")
	}

	objectLevel, ok := args[2].(float64)
	if !ok {
		return false, fmt.Errorf("levelMatch: objectLevel argument must be a number")
	}

	return levelMatch(action, subjectLevel, objectLevel), nil
}
