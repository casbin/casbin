package main

import (
	"github.com/lxmgo/config"
	"os"
	"strings"
	"fmt"
	"bufio"
	"io"
)

type Assertion struct {
	key string
	value string
	tokens []string
	policy [][]string
	rm *RoleManager
}

func (ast *Assertion) buildRoleLinks() {
	ast.rm = newRoleManager(1)
	for _, policy_role := range ast.policy {
		ast.rm.addLink(policy_role[0], policy_role[1])
	}
}

type Model map[string]AssertionMap
type AssertionMap map[string]*Assertion

func escape(s string) (string) {
	return strings.Replace(s, ".", "_", -1)
}

const (
	R_SECTION_NAME = "request_definition"
	P_SECTION_NAME = "policy_definition"
	G_SECTION_NAME = "role_definition"
	E_SECTION_NAME = "policy_effect"
	M_SECTION_NAME = "matchers"
)

var sectionNameMap = map[string]string {
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
	fmt.Println("Model:")
	for k, v := range model {
		for i, j := range v {
			fmt.Print(k + "." + i + ": ")
			fmt.Println(j.value)
		}
	}
}

func loadPolicy(path string, model Model) {
	fmt.Println("Policy:")
	readLine(path, model, loadPolicyLine)

	// model["g"].buildRoleLinks()
}

func loadPolicyLine(line string, model Model) {
	tokens := strings.Split(line, ", ")
	fmt.Println(tokens)

	key := tokens[0]
	sec := key[:1]
	model[sec][key].policy = append(model[sec][key].policy, tokens[1:])
}

func readLine(fileName string, model Model, handler func(string, Model)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line, model)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}
