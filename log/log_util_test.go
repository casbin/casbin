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
	"testing"

	"github.com/casbin/casbin/v2/log/mocks"

	"github.com/golang/mock/gomock"
)

func TestLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockLogger(ctrl)
	SetLogger(m)

	m.EXPECT().EnableLog(true)
	m.EXPECT().IsEnabled()

	logger.EnableLog(true)
	logger.IsEnabled()

	policy := map[string][][]string{}
	m.EXPECT().LogPolicy(policy)
	LogPolicy(policy)

	var model [][]string
	m.EXPECT().LogModel(model)
	LogModel(model)

	matcher := "my_matcher"
	request := []interface{}{"bob"}
	result := true
	var explains [][]string
	m.EXPECT().LogEnforce(matcher, request, result, explains)
	LogEnforce(matcher, request, result, explains)

	var roles []string
	m.EXPECT().LogRole(roles)
	LogRole(roles)
}
