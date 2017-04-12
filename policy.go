package casbin

import (
	"log"
	"strings"
	"os"
	"bufio"
	"io"
)

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

func getGroupingPolicy(model Model, ptype string) [][]string {
	return model["g"][ptype].policy
}
