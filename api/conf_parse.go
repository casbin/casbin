package api

import (
	"github.com/lxmgo/config"
)

type casbinConfig struct {
	modelPath string
	policyBackend string
	policyPath string
	dbDriver string
	dbDataSource string
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
