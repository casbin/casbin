package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
)

type Enforcer struct {
	modelPath string
	policyPath string
	model   Model
}

func (enforcer *Enforcer) init(modelPath string, policyPath string) {
	enforcer.modelPath = modelPath
	enforcer.policyPath = policyPath

	enforcer.reload()
}

func (enforcer *Enforcer) reload() {
	enforcer.model = loadModel(enforcer.modelPath)
	printModel(enforcer.model)

	loadPolicy(enforcer.policyPath, enforcer.model)
}

func (enforcer *Enforcer) enforce(rvals ...string) bool {
	fmt.Print("Request: ")
	fmt.Println(rvals)

	policyResults := make([]bool, len(enforcer.model["p"].policy))

	for i, pvals := range enforcer.model["p"].policy {
		//fmt.Print("Policy Rule: ")
		//fmt.Println(pvals)

		expression, _ := govaluate.NewEvaluableExpression(enforcer.model["m"].value)

		parameters := make(map[string]interface{}, 8)
		for j, token := range enforcer.model["r"].tokens {
			parameters[token] = rvals[j]
		}
		for j, token := range enforcer.model["p"].tokens {
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
	if enforcer.model["e"].value == "some(where (p.eft == allow))" {
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

