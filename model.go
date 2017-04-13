package casbin

import (
	"github.com/lxmgo/config"
	"log"
	"strings"
)

type Model map[string]AssertionMap
type AssertionMap map[string]*Assertion

func escape(s string) string {
	return strings.Replace(s, ".", "_", -1)
}

var sectionNameMap = map[string]string{
	"r": "request_definition",
	"p": "policy_definition",
	"g": "role_definition",
	"e": "policy_effect",
	"m": "matchers",
}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) {
	ast := Assertion{}
	ast.key = key
	ast.value = cfg.String(sectionNameMap[key] + "::" + key)

	if ast.value == "" {
		return
	}

	if sec == "r" || sec == "p" {
		ast.tokens = strings.Split(ast.value, ", ")
		for i := range ast.tokens {
			ast.tokens[i] = key + "_" + ast.tokens[i]
		}
	} else {
		ast.value = escape(ast.value)
	}

	_, ok := model[sec]
	if !ok {
		model[sec] = make(AssertionMap)
	}

	model[sec][key] = &ast
}

func loadModel(path string) (model Model) {
	cfg, _ := config.NewConfig(path)

	model = make(Model)

	loadAssertion(model, cfg, "r", "r")
	loadAssertion(model, cfg, "p", "p")
	loadAssertion(model, cfg, "e", "e")
	loadAssertion(model, cfg, "m", "m")

	loadAssertion(model, cfg, "g", "g")

	return model
}

func printModel(model Model) {
	log.Print("Model:")
	for k, v := range model {
		for i, j := range v {
			log.Printf("%s.%s: %s", k, i, j.value)
		}
	}
}
