package casbin

import (
	"github.com/lxmgo/config"
	"os"
	"strings"
	"bufio"
	"io"
	"log"
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

	log.Print("Role links for: " + ast.key)
	ast.rm.printRoles()
}

type Model map[string]AssertionMap
type AssertionMap map[string]*Assertion

func escape(s string) (string) {
	return strings.Replace(s, ".", "_", -1)
}

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
	log.Print("Model:")
	for k, v := range model {
		for i, j := range v {
			log.Printf("%s.%s: %s", k, i, j.value)
		}
	}
}

func loadPolicy(path string, model Model) {
	log.Print("Policy:")
	loadPolicyFile(path, model, loadPolicyLine)

	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}

	printPolicy(model)
}

func printPolicy(model Model) {
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}
}

func loadPolicyLine(line string, model Model) {
	tokens := strings.Split(line, ", ")

	key := tokens[0]
	sec := key[:1]
	model[sec][key].policy = append(model[sec][key].policy, tokens[1:])
}

func loadPolicyFile(fileName string, model Model, handler func(string, Model)) error {
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

func getPolicy(model Model, ptype string) [][]string {
	return model["p"][ptype].policy
}

func getFilteredPolicy(model Model, ptype string, fieldIndex int, fieldValue string) [][]string {
	res := [][]string{}

	for _, v := range model["p"][ptype].policy {
		if v[fieldIndex] == fieldValue {
			res = append(res, v)
		}
	}

	return res
}
