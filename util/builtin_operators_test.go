package util

import (
	"log"
	"testing"
)

func testKeyMatch(t *testing.T, key1 string, key2 string, res bool) {
	myRes := KeyMatch(key1, key2)
	log.Printf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch(t *testing.T) {
	testKeyMatch(t, "/foo", "/foo", true)
	testKeyMatch(t, "/foo", "/foo*", true)
	testKeyMatch(t, "/foo", "/foo/*", false)
	testKeyMatch(t, "/foo/bar", "/foo", false)
	testKeyMatch(t, "/foo/bar", "/foo*", true)
	testKeyMatch(t, "/foo/bar", "/foo/*", true)
	testKeyMatch(t, "/foobar", "/foo", false)
	testKeyMatch(t, "/foobar", "/foo*", true)
	testKeyMatch(t, "/foobar", "/foo/*", false)
}

func testRegexMatch(t *testing.T, key1 string, key2 string, res bool) {
	myRes := RegexMatch(key1, key2)
	log.Printf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestRegexMatch(t *testing.T) {
	testRegexMatch(t, "/topic/create", "/topic/create", true)
	testRegexMatch(t, "/topic/create/123", "/topic/create", true)
	testRegexMatch(t, "/topic/delete", "/topic/create", false)
	testRegexMatch(t, "/topic/edit", "/topic/edit/[0-9]+", false)
	testRegexMatch(t, "/topic/edit/123", "/topic/edit/[0-9]+", true)
	testRegexMatch(t, "/topic/edit/abc", "/topic/edit/[0-9]+", false)
	testRegexMatch(t, "/foo/delete/123", "/topic/delete/[0-9]+", false)
	testRegexMatch(t, "/topic/delete/0", "/topic/delete/[0-9]+", true)
	testRegexMatch(t, "/topic/edit/123s", "/topic/delete/[0-9]+", false)
}
