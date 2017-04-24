package util

import (
	"regexp"
	"strings"
)

// Determine whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, /foo/bar matches /foo/*
func KeyMatch(key1 string, key2 string) bool {
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

// The wrapper for KeyMatch.
func KeyMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return (bool)(KeyMatch(name1, name2)), nil
}

// Determine whether key1 matches the pattern of key2 in regular expression.
func RegexMatch(key1 string, key2 string) bool {
	res, err := regexp.MatchString(key2, key1)
	if err != nil {
		panic(err)
	}
	return res
}

// The wrapper for RegexMatch.
func RegexMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return (bool)(RegexMatch(name1, name2)), nil
}
