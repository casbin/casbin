package main

import (
	"fmt"
	"strings"
	"github.com/Knetic/govaluate"
)

type Model struct {
	r string
	p string
	e string
	m string
}

type Enforcer struct {
	model   Model
	rTokens []string
	pTokens []string
	policy  [][]string
}

func (enforcer *Enforcer) init(modelPath string, policyPath string) {
	enforcer.model = load_model(modelPath)
	fmt.Println("Model:")
	fmt.Println("r: " + enforcer.model.r)
	fmt.Println("p: " + enforcer.model.p)
	fmt.Println("e: " + enforcer.model.e)
	fmt.Println("m: " + enforcer.model.m)

	enforcer.rTokens = strings.Split(enforcer.model.r, ", ")
	for i := range enforcer.rTokens {
		enforcer.rTokens[i] = "r_" + enforcer.rTokens[i]
	}
	fmt.Println("R Tokens: ")
	fmt.Println(enforcer.rTokens)

	enforcer.pTokens = strings.Split(enforcer.model.p, ", ")
	for i := range enforcer.pTokens {
		enforcer.pTokens[i] = "p_" + enforcer.pTokens[i]
	}
	fmt.Println("P Tokens: ")
	fmt.Println(enforcer.pTokens)

	enforcer.policy = load_policy(policyPath)
	fmt.Println("Policy:")
	fmt.Println(enforcer.policy)
}

func (enforcer *Enforcer) enforce(rvals ...string) bool {
	fmt.Print("Request: ")
	fmt.Println(rvals)

	policyResults := make([]bool, len(enforcer.policy))

	for i, pvals := range enforcer.policy {
		//fmt.Print("Policy Rule: ")
		//fmt.Println(pvals)

		expression, _ := govaluate.NewEvaluableExpression(enforcer.model.m)

		parameters := make(map[string]interface{}, 8)
		for j, token := range enforcer.rTokens {
			parameters[token] = rvals[j]
		}
		for j, token := range enforcer.pTokens {
			parameters[token] = pvals[j]
		}

		result, _ := expression.Evaluate(parameters)
		//fmt.Print("Result: ")
		//fmt.Println(result)

		policyResults[i] = result.(bool)
	}

	//fmt.Print("Rule Results: ")
	//fmt.Println(policyResults)

	result := false
	if enforcer.model.e == "some(where (p.eft == allow))" {
		result = false
		for _, res := range policyResults {
			if res {
				result = true
				break
			}
		}
	}

	fmt.Print("Final Result: ")
	fmt.Println(result)

	return result
}

