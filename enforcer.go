package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"strings"
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

func (enforcer *Enforcer) keyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	} else {
		if len(key1) > i {
			return key1[:i] == key2[:i]
		} else {
			return key1 == key2[:i]
		}
	}
}

func (enforcer *Enforcer) enforce(rvals ...string) bool {
	fmt.Print("Request ")
	fmt.Print(rvals)

	expString := enforcer.model["m"]["m"].value
	var expression *govaluate.EvaluableExpression = nil

	_, ok := enforcer.model["g"]
	if !ok {
		expression, _ = govaluate.NewEvaluableExpression(expString)
	} else {
		functions := make(map[string]govaluate.ExpressionFunction)

		for key, ast := range enforcer.model["g"] {
			functions[key] = func(args ...interface{}) (interface{}, error) {
				name1 := args[0].(string)
				name2 := args[1].(string)

				return (bool)(ast.rm.hasLink(name1, name2)), nil
			}
		}

		expression, _ = govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
	}

	policyResults := make([]bool, len(enforcer.model["p"]["p"].policy))

	for i, pvals := range enforcer.model["p"]["p"].policy {
		//fmt.Print("Policy Rule: ")
		//fmt.Println(pvals)

		parameters := make(map[string]interface{}, 8)
		for j, token := range enforcer.model["r"]["r"].tokens {
			parameters[token] = rvals[j]
		}
		for j, token := range enforcer.model["p"]["p"].tokens {
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
	if enforcer.model["e"]["e"].value == "some(where (p_eft == allow))" {
		result = false
		for _, res := range policyResults {
			if res {
				result = true
				break
			}
		}
	}

	fmt.Print(": ")
	fmt.Println(result)

	return result
}

