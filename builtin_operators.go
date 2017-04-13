package casbin

import "strings"

func keyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	} else {
		return key1 == key2[:i]
	}
}
