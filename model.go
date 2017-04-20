package casbin

import (
	"github.com/hsluoyz/casbin/util"
	"github.com/lxmgo/config"
	"log"
	"strconv"
	"strings"
)

// Model represents the whole access control model.
type Model map[string]AssertionMap

// AssertionMap is the collection of assertions, can be "r", "p", "g", "e", "m".
type AssertionMap map[string]*Assertion

var sectionNameMap = map[string]string{
	"r": "request_definition",
	"p": "policy_definition",
	"g": "role_definition",
	"e": "policy_effect",
	"m": "matchers",
}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) bool {
	ast := Assertion{}
	ast.Key = key
	ast.Value = cfg.String(sectionNameMap[sec] + "::" + key)

	if ast.Value == "" {
		return false
	}

	if sec == "m" {
		ast.Value = util.FixAttribute(ast.Value)
	}

	if sec == "r" || sec == "p" {
		ast.Tokens = strings.Split(ast.Value, ", ")
		for i := range ast.Tokens {
			ast.Tokens[i] = key + "_" + ast.Tokens[i]
		}
	} else {
		ast.Value = util.EscapeAssertion(ast.Value)
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
		if !loadAssertion(model, cfg, sec, sec+getKeySuffix(i)) {
			break
		} else {
			i++
		}
	}
}

// Load the model from model CONF file.
func LoadModel(path string) Model {
	cfg, err := config.NewConfig(path)
	if err != nil {
		panic(err)
	}

	model := make(Model)

	loadSection(model, cfg, "r")
	loadSection(model, cfg, "p")
	loadSection(model, cfg, "e")
	loadSection(model, cfg, "m")

	loadSection(model, cfg, "g")

	return model
}

// Print the model.
func (model Model) PrintModel() {
	log.Print("Model:")
	for k, v := range model {
		for i, j := range v {
			log.Printf("%s.%s: %s", k, i, j.Value)
		}
	}
}
