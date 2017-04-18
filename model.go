package casbin

import (
	"github.com/lxmgo/config"
	"log"
	"strconv"
	"strings"
	"github.com/hsluoyz/casbin/util"
)

// Model represents the whole access control model.
type Model map[string]AssertionMap

// AssertionMap is the collection of assertions, can be "r", "p", "g", "e", "m".
type AssertionMap map[string]*assertion

var sectionNameMap = map[string]string{
	"r": "request_definition",
	"p": "policy_definition",
	"g": "role_definition",
	"e": "policy_effect",
	"m": "matchers",
}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) bool {
	ast := assertion{}
	ast.key = key
	ast.value = cfg.String(sectionNameMap[sec] + "::" + key)

	if ast.value == "" {
		return false
	}

	if sec == "m" {
		ast.value = util.FixAttribute(ast.value)
	}

	if sec == "r" || sec == "p" {
		ast.tokens = strings.Split(ast.value, ", ")
		for i := range ast.tokens {
			ast.tokens[i] = key + "_" + ast.tokens[i]
		}
	} else {
		ast.value = util.EscapeAssertion(ast.value)
	}

	_, ok := model[sec]
	if !ok {
		model[sec] = make(AssertionMap)
	}

	model[sec][key] = &ast
	return true
}

func getKeySuffix(i int) string {
	if i == 1 {
		return ""
	} else {
		return strconv.Itoa(i)
	}
}

func loadSection(model Model, cfg config.ConfigInterface, sec string) {
	i := 1
	for {
		if !loadAssertion(model, cfg, sec, sec + getKeySuffix(i)) {
			break
		} else {
			i++
		}
	}
}

func loadModel(path string) Model {
	cfg, _ := config.NewConfig(path)

	model := make(Model)

	loadSection(model, cfg, "r")
	loadSection(model, cfg, "p")
	loadSection(model, cfg, "e")
	loadSection(model, cfg, "m")

	loadSection(model, cfg, "g")

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
