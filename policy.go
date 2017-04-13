package casbin

import (
	"log"
	"strings"
	"os"
	"bufio"
	"io"
	"bytes"
)

func buildRoleLinks(model Model) {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

func loadPolicy(path string, model Model) {
	clearPolicy(model)
	log.Print("Policy:")
	loadPolicyFile(path, model, loadPolicyLine)
}

func printPolicy(model Model) {
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.value, ": ", ast.policy)
	}
}

func arrayToString(s []string) string {
	var tmp bytes.Buffer
	for i, v := range s {
		if i != len(s) - 1 {
			tmp.WriteString(v + ", ")
		} else {
			tmp.WriteString(v)
		}
	}
	return tmp.String()
}

func savePolicy(path string, model Model) {
	var tmp bytes.Buffer

	for key, ast := range model["p"] {
		for _, rule := range ast.policy {
			tmp.WriteString(key + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for key, ast := range model["g"] {
		for _, rule := range ast.policy {
			tmp.WriteString(key + ", ")
			tmp.WriteString(arrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	savePolicyFile(path, strings.TrimRight(tmp.String(), "\n"))
}

func clearPolicy(model Model) {
	for _, ast := range model["p"] {
		ast.policy = nil
	}

	for _, ast := range model["g"] {
		ast.policy = nil
	}
}

func loadPolicyLine(line string, model Model) {
	if line == "" {
		return
	}

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

func savePolicyFile(fileName string, text string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	w.WriteString(text)
	w.Flush()
	f.Close()
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
