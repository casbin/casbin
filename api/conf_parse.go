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

package api

import (
	"github.com/lxmgo/config"
)

type casbinConfig struct {
	modelPath     string
	policyBackend string
	policyPath    string
	dbDriver      string
	dbDataSource  string
}

func loadConfig(cfgPath string) *casbinConfig {
	ccfg := casbinConfig{}
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		panic(err)
	}

	ccfg.modelPath = cfg.String("default::model_path")
	ccfg.policyBackend = cfg.String("default::policy_backend")

	if ccfg.policyBackend == "file" {
		ccfg.policyPath = cfg.String("file::policy_path")
	} else if ccfg.policyBackend == "database" {
		ccfg.dbDriver = cfg.String("database::driver")
		ccfg.dbDataSource = cfg.String("database::data_source")
	}

	return &ccfg
}
