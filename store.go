package main

import (
	"github.com/lxmgo/config"
	"os"
	"io/ioutil"
	"strings"
)

func load_model(path string) (model Model) {
	config, _ := config.NewConfig(path)
	model = Model{}
	model.r = config.String("request_definition::r")
	model.p = config.String("policy_definition::p")
	model.e = config.String("policy_effect::e")
	model.m = strings.Replace(config.String("matchers::m"), ".", "_", -1)
	return model
}

func load_policy(path string) ([][]string) {
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
