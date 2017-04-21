package util

import (
	"regexp"
	"strings"
)

// Escape the dots in the assertion, because the expression evaluation doesn't support such variable names.
func EscapeAssertion(s string) string {
	return strings.Replace(s, ".", "_", -1)
}

// Translate the ABAC attributes into functions.
func FixAttribute(s string) string {
	reg := regexp.MustCompile("r\\.sub\\.([A-Za-z0-9]*)")
	res := reg.ReplaceAllString(s, "subAttr(r.sub, \"$1\")")

	reg = regexp.MustCompile("r\\.obj\\.([A-Za-z0-9]*)")
	res = reg.ReplaceAllString(res, "objAttr(r.obj, \"$1\")")

	return res
}

// Determine whether two string arrays are identical.
func ArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Determine whether two 2-dimensional string arrays are identical.
func Array2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !ArrayEquals(v, b[i]) {
			return false
		}
	}
	return true
}

// Remove any duplicated elements in a string array.
func ArrayRemoveDuplicates(s *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *s {
		if !found[x] {
			found[x] = true
			(*s)[j] = (*s)[i]
			j++
		}
	}
	*s = (*s)[:j]
}

// Get a printable string for a string array.
func ArrayToString(s []string) string {
	return strings.Join(s, ", ")
}
