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
