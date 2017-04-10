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
}

type Model map[string]*Assertion

func escape(s string) (string) {
	return strings.Replace(s, ".", "_", -1)
}

func loadAssertion(cfg config.ConfigInterface, sec string, key string) (*Assertion) {
	ast := Assertion{}
	ast.key = key
	ast.value = cfg.String(sec + "::" + key)

	if sec == "request_definition" || sec == "policy_definition" {
		ast.tokens = strings.Split(ast.value, ", ")
		for i := range ast.tokens {
			ast.tokens[i] = key + "_" + ast.tokens[i]
		}
	} else if sec == "matchers" {
		ast.value = escape(ast.value)
	}

	return &ast
}

func loadModel(path string) (model Model) {
	cfg, _ := config.NewConfig(path)

	model = make(Model)

	model["r"] = loadAssertion(cfg, "request_definition", "r")
	model["p"] = loadAssertion(cfg, "policy_definition", "p")
	model["e"] = loadAssertion(cfg, "policy_effect", "e")
	model["m"] = loadAssertion(cfg, "matchers", "m")

	//fmt.Println("R Tokens: ")
	//fmt.Println(model["r"].tokens)
	//
	//fmt.Println("P Tokens: ")
	//fmt.Println(model["p"].tokens)

	return model
}

func loadPolicy(path string, model Model) {
	fmt.Println("Policy: ")
	readLine(path, model, loadPolicyLine)
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
