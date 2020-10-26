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

import (
	"log"
	"sync/atomic"
)

// DefaultLogger is the implementation for a Logger using golang log.
type DefaultLogger struct {
	enable int32
}

func (l *DefaultLogger) EnableLog(enable bool) {
	i := 0
	if enable {
		i = 1
	}
	atomic.StoreInt32(&(l.enable), int32(i))
}

func (l *DefaultLogger) IsEnabled() bool {
	return atomic.LoadInt32(&(l.enable)) != 0
}

func (l *DefaultLogger) LogModel(event int, line []string, model [][]string) {
	for _, v := range line {
		log.Print(v)
	}
}

func (l *DefaultLogger) LogEnforce(event int, line string, request *[]interface{}, policies *[]string, result *[]interface{}) {
	log.Print(line)
}

func (l *DefaultLogger) LogPolicy(event int, line string, pPolicyFormat []string, gPolicyFormat []string, pPolicy *[]interface{}, gPolicy *[]interface{}) {
	log.Print(line)
	if pPolicy != nil {
		for k, v := range *pPolicy {
			log.Print("p: ", pPolicyFormat[k], ": ", v)
		}
	}
	if gPolicy != nil {
		for k, v := range *gPolicy {
			log.Print("g: ", gPolicyFormat[k], ": ", v)
		}
	}
}

func (l *DefaultLogger) LogRole(event int, line string, role []string) {
	log.Print(line)
}

/*
func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	if l.IsEnabled() {
		log.Printf(format, v...)
	}
}
*/
