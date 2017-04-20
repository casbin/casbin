package persist

import (
	"strings"
	"github.com/hsluoyz/casbin"
)

func loadPolicyLine(line string, model casbin.Model) {
	if line == "" {
		return
	}

	tokens := strings.Split(line, ", ")

	key := tokens[0]
	sec := key[:1]
	model[sec][key].Policy = append(model[sec][key].Policy, tokens[1:])
}

// The abstract adapter interface for policy persistence.
// FileAdapter, DBAdapter inherits this interface.
type Adapter interface {
	LoadPolicy(model casbin.Model)
	SavePolicy(model casbin.Model)
}
