// Copyright 2018 The casbin Authors. All Rights Reserved.
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

//go:generate mockgen -destination=./mocks/mock_logger.go -package=mocks github.com/casbin/casbin/v2/log Logger

// Logger is the logging interface implementation.
type Logger interface {
	// EnableLog controls whether print the message.
	EnableLog(bool)

	// IsEnabled returns if logger is enabled.
	IsEnabled() bool

	// LogModel log info related to model.
	LogModel(model [][]string)

	// LogEnforce log info related to enforce.
	LogEnforce(matcher string, request []interface{}, result bool, explains [][]string)

	// LogRole log info related to role.
	LogRole(roles []string)

	// LogPolicy log info related to policy.
	LogPolicy(policy map[string][][]string)
}
