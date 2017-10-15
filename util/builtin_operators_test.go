// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"log"
	"testing"
)

func testKeyMatch(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
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

func testKeyMatch2(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch2(key1, key2)
	log.Printf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch2(t *testing.T) {
	testKeyMatch2(t, "/foo", "/foo", true)
	testKeyMatch2(t, "/foo", "/foo*", true)
	testKeyMatch2(t, "/foo", "/foo/*", false)
	testKeyMatch2(t, "/foo/bar", "/foo", true) // different with KeyMatch.
	testKeyMatch2(t, "/foo/bar", "/foo*", true)
	testKeyMatch2(t, "/foo/bar", "/foo/*", true)
	testKeyMatch2(t, "/foobar", "/foo", true) // different with KeyMatch.
	testKeyMatch2(t, "/foobar", "/foo*", true)
	testKeyMatch2(t, "/foobar", "/foo/*", false)

	testKeyMatch2(t, "/", "/:resource", false)
	testKeyMatch2(t, "/resource1", "/:resource", true)
	testKeyMatch2(t, "/myid", "/:id/using/:resId", false)
	testKeyMatch2(t, "/myid/using/myresid", "/:id/using/:resId", true)

	testKeyMatch2(t, "/proxy/myid", "/proxy/:id/*", false)
	testKeyMatch2(t, "/proxy/myid/", "/proxy/:id/*", true)
	testKeyMatch2(t, "/proxy/myid/res", "/proxy/:id/*", true)
	testKeyMatch2(t, "/proxy/myid/res/res2", "/proxy/:id/*", true)
	testKeyMatch2(t, "/proxy/myid/res/res2/res3", "/proxy/:id/*", true)
	testKeyMatch2(t, "/proxy/", "/proxy/:id/*", false)

	testKeyMatch2(t, "/alice", "/:id", true)
	testKeyMatch2(t, "/alice/all", "/:id/all", true)
	testKeyMatch2(t, "/alice", "/:id/all", false)
	testKeyMatch2(t, "/alice/all", "/:id", false)
}

func testKeyMatch3(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch3(key1, key2)
	log.Printf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch3(t *testing.T) {
	testKeyMatch3(t, "/foo", "/foo", true)
	testKeyMatch3(t, "/foo", "/foo*", true)
	testKeyMatch3(t, "/foo", "/foo/*", false)
	testKeyMatch3(t, "/foo/bar", "/foo", true) // different with KeyMatch.
	testKeyMatch3(t, "/foo/bar", "/foo*", true)
	testKeyMatch3(t, "/foo/bar", "/foo/*", true)
	testKeyMatch3(t, "/foobar", "/foo", true) // different with KeyMatch.
	testKeyMatch3(t, "/foobar", "/foo*", true)
	testKeyMatch3(t, "/foobar", "/foo/*", false)

	testKeyMatch3(t, "/", "/{resource}", false)
	testKeyMatch3(t, "/resource1", "/{resource}", true)
	testKeyMatch3(t, "/myid", "/{id}/using/{resId}", false)
	testKeyMatch3(t, "/myid/using/myresid", "/{id}/using/{resId}", true)

	testKeyMatch3(t, "/proxy/myid", "/proxy/{id}/*", false)
	testKeyMatch3(t, "/proxy/myid/", "/proxy/{id}/*", true)
	testKeyMatch3(t, "/proxy/myid/res", "/proxy/{id}/*", true)
	testKeyMatch3(t, "/proxy/myid/res/res2", "/proxy/{id}/*", true)
	testKeyMatch3(t, "/proxy/myid/res/res2/res3", "/proxy/{id}/*", true)
	testKeyMatch3(t, "/proxy/", "/proxy/{id}/*", false)
}

func testRegexMatch(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
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

func testIPMatch(t *testing.T, ip1 string, ip2 string, res bool) {
	t.Helper()
	myRes := IPMatch(ip1, ip2)
	log.Printf("%s < %s: %t", ip1, ip2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", ip1, ip2, !res, res)
	}
}

func TestIPMatch(t *testing.T) {
	testIPMatch(t, "192.168.2.123", "192.168.2.0/24", true)
	testIPMatch(t, "192.168.2.123", "192.168.3.0/24", false)
	testIPMatch(t, "192.168.2.123", "192.168.2.0/16", true)
	testIPMatch(t, "192.168.2.123", "192.168.2.123", true)
	testIPMatch(t, "192.168.2.123", "192.168.2.123/32", true)
	testIPMatch(t, "10.0.0.11", "10.0.0.0/8", true)
	testIPMatch(t, "11.0.0.123", "10.0.0.0/8", false)
}
