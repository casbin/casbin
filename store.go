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

type Model map[string]*Assertion

func escape(s string) (string) {
	return strings.Replace(s, ".", "_", -1)
}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) {
	ast := Assertion{}
	ast.key = key
	ast.value = cfg.String(sec + "::" + key)

	if ast.value == "" {
		return
	}

	if sec == "request_definition" || sec == "policy_definition" {
		ast.tokens = strings.Split(ast.value, ", ")
		for i := range ast.tokens {
			ast.tokens[i] = key + "_" + ast.tokens[i]
		}
	} else {
		ast.value = escape(ast.value)
	}

	model[key] = &ast
}

func loadModel(path string) (model Model) {
	cfg, _ := config.NewConfig(path)

	model = make(Model)

	loadAssertion(model, cfg, "request_definition", "r")
	loadAssertion(model, cfg, "policy_definition", "p")
	loadAssertion(model, cfg, "policy_effect", "e")
	loadAssertion(model, cfg, "matchers", "m")

	loadAssertion(model, cfg, "role_definition", "g")

	return model
}

func printModel(model Model) {
	fmt.Println("Model:")
	for k, v := range model {
		fmt.Println(k + ": " + v.value)
	}
}

func loadPolicy(path string, model Model) {
	fmt.Println("Policy:")
	readLine(path, model, loadPolicyLine)

	model["g"].buildRoleLinks()
}

func loadPolicyLine(line string, model Model) {
	tokens := strings.Split(line, ", ")
	fmt.Println(tokens)

	model[tokens[0]].policy = append(model[tokens[0]].policy, tokens[1:])
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
