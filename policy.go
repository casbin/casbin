package casbin

import (
	"github.com/hsluoyz/casbin/util"
	"log"
	"strings"
)

func buildRoleLinks(model Model) {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

func printPolicy(model Model) {
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
	}
}

func clearPolicy(model Model) {
	for _, ast := range model["p"] {
		ast.Policy = nil
	}

	for _, ast := range model["g"] {
		ast.Policy = nil
	}
}

func loadPolicyLine(line string, model Model) {
	if line == "" {
		return
	}

	tokens := strings.Split(line, ", ")

	key := tokens[0]
	sec := key[:1]
	model[sec][key].Policy = append(model[sec][key].Policy, tokens[1:])
}

func getPolicy(model Model, sec string, ptype string) [][]string {
	return model[sec][ptype].Policy
}

func getFilteredPolicy(model Model, sec string, ptype string, fieldIndex int, fieldValue string) [][]string {
	res := [][]string{}

	for _, v := range model[sec][ptype].Policy {
		if v[fieldIndex] == fieldValue {
			res = append(res, v)
		}
	}

	return res
}

func hasPolicy(model Model, sec string, ptype string, policy []string) bool {
	for _, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			return true
		}
	}

	return false
}

func addPolicy(model Model, sec string, ptype string, policy []string) bool {
	if !hasPolicy(model, sec, ptype, policy) {
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, policy)
		return true
	} else {
		return false
	}
}

func removePolicy(model Model, sec string, ptype string, policy []string) bool {
	for i, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			return true
		}
	}

	return false
}

func getValuesForFieldInPolicy(model Model, sec string, ptype string, fieldIndex int) []string {
	users := []string{}

	for _, rule := range model[sec][ptype].Policy {
		users = append(users, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&users)
	// sort.Strings(users)

	return users
}
