package main

import (
	"github.com/lxmgo/config"
	"os"
	"io/ioutil"
	"strings"
	"fmt"
)

type Model struct {
	r Assertion
	p Assertion
	e Assertion
	m Assertion
}

type Assertion struct {
	key string
	value string
	tokens []string
}

func escape(s string) (string) {
	return strings.Replace(s, ".", "_", -1)
}

func loadAssertion(cfg config.ConfigInterface, sec string, key string) (Assertion) {
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

	return ast
}

func loadModel(path string) (model Model) {
	cfg, _ := config.NewConfig(path)
	model = Model{}

	model.r = loadAssertion(cfg, "request_definition", "r")
	model.p = loadAssertion(cfg, "policy_definition", "p")
	model.e = loadAssertion(cfg, "policy_effect", "e")
	model.m = loadAssertion(cfg, "matchers", "m")

	fmt.Println("R Tokens: ")
	fmt.Println(model.r.tokens)

	fmt.Println("P Tokens: ")
	fmt.Println(model.p.tokens)

	return model
}

func loadPolicy(path string) ([][]string) {
	fi, err := os.Open(path)
	if err != nil{panic(err)}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	text := string(fd)

	column := 0
	lines := strings.Split(text, "\r\n")
	row := len(lines)
	if row > 0 {
		policyLine := strings.Split(lines[0], ", ")
		column = len(policyLine)
	}

	if column == 0 {
		return nil
	}

	policyLines := make([][]string, row)

	for i, line := range lines {
		policyLine := strings.Split(line, ", ")
		policyLines[i] = policyLine
	}

	return policyLines
}
