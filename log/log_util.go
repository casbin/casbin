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

var logger Logger = &DefaultLogger{}

// SetLogger sets the current logger.
func SetLogger(l Logger) {
	logger = l
}

// GetLogger returns the current logger.
func GetLogger() Logger {
	return logger
}

/*
// LogPrint prints the log.
func LogPrint(v ...interface{}) {
	logger.Print(v...)
}

// LogPrintf prints the log with the format.
func LogPrintf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}
*/

// LogModel logs the model information.
func LogModel(event int, line []string, model [][]string) {
	logger.LogModel(event, line, model)
}

// LogEnforce logs the enforcer information.
func LogEnforce(event int, line string, request *[]interface{}, policies *[]string, result *[]interface{}) {
	logger.LogEnforce(event, line, request, policies, result)
}

// LogRole logs the role information.
func LogRole(event int, line string, role []string) {
	logger.LogRole(event, line, role)
}

// LogPolicy logs the policy information.
func LogPolicy(event int, line string, pPolicyFormat []string, gPolicyFormat []string, pPolicy *[]interface{}, gPolicy *[]interface{}) {
	logger.LogPolicy(event, line, pPolicyFormat, gPolicyFormat, pPolicy, gPolicy)
}
