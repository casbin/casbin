package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
)

type Enforcer struct {
	modelPath string
	policyPath string
	model   Model
	policy  [][]string
}

func (enforcer *Enforcer) init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath

	enforcer.reload()
}

func (enforcer *Enforcer) reload() {
	enforcer.model = loadModel(enforcer.modelPath)
	fmt.Println("Model:")
	fmt.Println("r: " + enforcer.model.r.value)
	fmt.Println("p: " + enforcer.model.p.value)
	fmt.Println("e: " + enforcer.model.e.value)
	fmt.Println("m: " + enforcer.model.m.value)

	enforcer.policy = loadPolicy(enforcer.policyPath)
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

		expression, _ := govaluate.NewEvaluableExpression(enforcer.model.m.value)

		parameters := make(map[string]interface{}, 8)
		for j, token := range enforcer.model.r.tokens {
			parameters[token] = rvals[j]
		}
		for j, token := range enforcer.model.p.tokens {
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
	if enforcer.model.e.value == "some(where (p.eft == allow))" {
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

