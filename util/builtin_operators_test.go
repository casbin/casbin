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
	"testing"
)

func testKeyMatch(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch(key1, key2)
	t.Logf("%s < %s: %t", key1, key2, myRes)

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

func testKeyGet(t *testing.T, key1 string, key2 string, res string) {
	t.Helper()
	myRes := KeyGet(key1, key2)
	t.Logf(`%s < %s: "%s"`, key1, key2, myRes)

	if myRes != res {
		t.Errorf(`%s < %s: "%s", supposed to be "%s"`, key1, key2, myRes, res)
	}
}

func TestKeyGet(t *testing.T) {
	testKeyGet(t, "/foo", "/foo", "")
	testKeyGet(t, "/foo", "/foo*", "")
	testKeyGet(t, "/foo", "/foo/*", "")
	testKeyGet(t, "/foo/bar", "/foo", "")
	testKeyGet(t, "/foo/bar", "/foo*", "/bar")
	testKeyGet(t, "/foo/bar", "/foo/*", "bar")
	testKeyGet(t, "/foobar", "/foo", "")
	testKeyGet(t, "/foobar", "/foo*", "bar")
	testKeyGet(t, "/foobar", "/foo/*", "")
}

func testKeyMatch2(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch2(key1, key2)
	t.Logf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func testGlobMatch(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes, err := GlobMatch(key1, key2)
	if err != nil {
		panic(err)
	}
	t.Logf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch2(t *testing.T) {
	testKeyMatch2(t, "/foo", "/foo", true)
	testKeyMatch2(t, "/foo", "/foo*", true)
	testKeyMatch2(t, "/foo", "/foo/*", false)
	testKeyMatch2(t, "/foo/bar", "/foo", false)
	testKeyMatch2(t, "/foo/bar", "/foo*", false) // different with KeyMatch.
	testKeyMatch2(t, "/foo/bar", "/foo/*", true)
	testKeyMatch2(t, "/foobar", "/foo", false)
	testKeyMatch2(t, "/foobar", "/foo*", false) // different with KeyMatch.
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

	testKeyMatch2(t, "/alice/all", "/:/all", false)
}

func testKeyGet2(t *testing.T, key1 string, key2 string, pathVar string, res string) {
	t.Helper()
	myRes := KeyGet2(key1, key2, pathVar)
	t.Logf(`%s < %s: %s = "%s"`, key1, key2, pathVar, myRes)

	if myRes != res {
		t.Errorf(`%s < %s: %s = "%s" supposed to be "%s"`, key1, key2, pathVar, myRes, res)
	}
}

func TestKeyGet2(t *testing.T) {
	testKeyGet2(t, "/foo", "/foo", "id", "")
	testKeyGet2(t, "/foo", "/foo*", "id", "")
	testKeyGet2(t, "/foo", "/foo/*", "id", "")
	testKeyGet2(t, "/foo/bar", "/foo", "id", "")
	testKeyGet2(t, "/foo/bar", "/foo*", "id", "") // different with KeyMatch.
	testKeyGet2(t, "/foo/bar", "/foo/*", "id", "")
	testKeyGet2(t, "/foobar", "/foo", "id", "")
	testKeyGet2(t, "/foobar", "/foo*", "id", "") // different with KeyMatch.
	testKeyGet2(t, "/foobar", "/foo/*", "id", "")

	testKeyGet2(t, "/", "/:resource", "resource", "")
	testKeyGet2(t, "/resource1", "/:resource", "resource", "resource1")
	testKeyGet2(t, "/myid", "/:id/using/:resId", "id", "")
	testKeyGet2(t, "/myid/using/myresid", "/:id/using/:resId", "id", "myid")
	testKeyGet2(t, "/myid/using/myresid", "/:id/using/:resId", "resId", "myresid")

	testKeyGet2(t, "/proxy/myid", "/proxy/:id/*", "id", "")
	testKeyGet2(t, "/proxy/myid/", "/proxy/:id/*", "id", "myid")
	testKeyGet2(t, "/proxy/myid/res", "/proxy/:id/*", "id", "myid")
	testKeyGet2(t, "/proxy/myid/res/res2", "/proxy/:id/*", "id", "myid")
	testKeyGet2(t, "/proxy/myid/res/res2/res3", "/proxy/:id/*", "id", "myid")
	testKeyGet2(t, "/proxy/myid/res/res2/res3", "/proxy/:id/res/*", "id", "myid")
	testKeyGet2(t, "/proxy/", "/proxy/:id/*", "id", "")

	testKeyGet2(t, "/alice", "/:id", "id", "alice")
	testKeyGet2(t, "/alice/all", "/:id/all", "id", "alice")
	testKeyGet2(t, "/alice", "/:id/all", "id", "")
	testKeyGet2(t, "/alice/all", "/:id", "id", "")

	testKeyGet2(t, "/alice/all", "/:/all", "", "")
}
func testKeyMatch3(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch3(key1, key2)
	t.Logf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch3(t *testing.T) {
	// keyMatch3() is similar with KeyMatch2(), except using "/proxy/{id}" instead of "/proxy/:id".
	testKeyMatch3(t, "/foo", "/foo", true)
	testKeyMatch3(t, "/foo", "/foo*", true)
	testKeyMatch3(t, "/foo", "/foo/*", false)
	testKeyMatch3(t, "/foo/bar", "/foo", false)
	testKeyMatch3(t, "/foo/bar", "/foo*", false)
	testKeyMatch3(t, "/foo/bar", "/foo/*", true)
	testKeyMatch3(t, "/foobar", "/foo", false)
	testKeyMatch3(t, "/foobar", "/foo*", false)
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

	testKeyMatch3(t, "/myid/using/myresid", "/{id/using/{resId}", false)
}

func testKeyMatch4(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := KeyMatch4(key1, key2)
	t.Logf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch4(t *testing.T) {
	testKeyMatch4(t, "/parent/123/child/123", "/parent/{id}/child/{id}", true)
	testKeyMatch4(t, "/parent/123/child/456", "/parent/{id}/child/{id}", false)

	testKeyMatch4(t, "/parent/123/child/123", "/parent/{id}/child/{another_id}", true)
	testKeyMatch4(t, "/parent/123/child/456", "/parent/{id}/child/{another_id}", true)

	testKeyMatch4(t, "/parent/123/child/123/book/123", "/parent/{id}/child/{id}/book/{id}", true)
	testKeyMatch4(t, "/parent/123/child/123/book/456", "/parent/{id}/child/{id}/book/{id}", false)
	testKeyMatch4(t, "/parent/123/child/456/book/123", "/parent/{id}/child/{id}/book/{id}", false)
	testKeyMatch4(t, "/parent/123/child/456/book/", "/parent/{id}/child/{id}/book/{id}", false)
	testKeyMatch4(t, "/parent/123/child/456", "/parent/{id}/child/{id}/book/{id}", false)

	testKeyMatch4(t, "/parent/123/child/123", "/parent/{i/d}/child/{i/d}", false)
}

func testRegexMatch(t *testing.T, key1 string, key2 string, res bool) {
	t.Helper()
	myRes := RegexMatch(key1, key2)
	t.Logf("%s < %s: %t", key1, key2, myRes)

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
	t.Logf("%s < %s: %t", ip1, ip2, myRes)

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

func testRegexMatchFunc(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := RegexMatchFunc(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func testKeyMatchFunc(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := KeyMatchFunc(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func testKeyMatch2Func(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := KeyMatch2Func(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func testKeyMatch3Func(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := KeyMatch3Func(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func testKeyMatch4Func(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := KeyMatch4Func(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func testIPMatchFunc(t *testing.T, res bool, err string, args ...interface{}) {
	t.Helper()
	myRes, myErr := IPMatchFunc(args...)
	myErrStr := ""

	if myErr != nil {
		myErrStr = myErr.Error()
	}

	if myRes != res || err != myErrStr {
		t.Errorf("%v returns %v %v, supposed to be %v %v", args, myRes, myErr, res, err)
	}
}

func TestRegexMatchFunc(t *testing.T) {
	testRegexMatchFunc(t, false, "regexMatch: Expected 2 arguments, but got 1", "/topic/create")
	testRegexMatchFunc(t, false, "regexMatch: Expected 2 arguments, but got 3", "/topic/create/123", "/topic/create", "/topic/update")
	testRegexMatchFunc(t, false, "regexMatch: Argument must be a string", "/topic/create", false)
	testRegexMatchFunc(t, true, "", "/topic/create/123", "/topic/create")
}

func TestKeyMatchFunc(t *testing.T) {
	testKeyMatchFunc(t, false, "keyMatch: Expected 2 arguments, but got 1", "/foo")
	testKeyMatchFunc(t, false, "keyMatch: Expected 2 arguments, but got 3", "/foo/create/123", "/foo/*", "/foo/update/123")
	testKeyMatchFunc(t, false, "keyMatch: Argument must be a string", "/foo", true)
	testKeyMatchFunc(t, false, "", "/foo/bar", "/foo")
	testKeyMatchFunc(t, true, "", "/foo/bar", "/foo/*")
	testKeyMatchFunc(t, true, "", "/foo/bar", "/foo*")
}

func TestKeyMatch2Func(t *testing.T) {
	testKeyMatch2Func(t, false, "keyMatch2: Expected 2 arguments, but got 1", "/")
	testKeyMatch2Func(t, false, "keyMatch2: Expected 2 arguments, but got 3", "/foo/create/123", "/*", "/foo/update/123")
	testKeyMatch2Func(t, false, "keyMatch2: Argument must be a string", "/foo", true)

	testKeyMatch2Func(t, false, "", "/", "/:resource")
	testKeyMatch2Func(t, true, "", "/resource1", "/:resource")

	testKeyMatch2Func(t, true, "", "/foo", "/foo")
	testKeyMatch2Func(t, true, "", "/foo", "/foo*")
	testKeyMatch2Func(t, false, "", "/foo", "/foo/*")
}

func TestKeyMatch3Func(t *testing.T) {
	testKeyMatch3Func(t, false, "keyMatch3: Expected 2 arguments, but got 1", "/")
	testKeyMatch3Func(t, false, "keyMatch3: Expected 2 arguments, but got 3", "/foo/create/123", "/*", "/foo/update/123")
	testKeyMatch3Func(t, false, "keyMatch3: Argument must be a string", "/foo", true)

	testKeyMatch3Func(t, true, "", "/foo", "/foo")
	testKeyMatch3Func(t, true, "", "/foo", "/foo*")
	testKeyMatch3Func(t, false, "", "/foo", "/foo/*")
	testKeyMatch3Func(t, false, "", "/foo/bar", "/foo")
	testKeyMatch3Func(t, false, "", "/foo/bar", "/foo*")
	testKeyMatch3Func(t, true, "", "/foo/bar", "/foo/*")
	testKeyMatch3Func(t, false, "", "/foobar", "/foo")
	testKeyMatch3Func(t, false, "", "/foobar", "/foo*")
	testKeyMatch3Func(t, false, "", "/foobar", "/foo/*")

	testKeyMatch3Func(t, false, "", "/", "/{resource}")
	testKeyMatch3Func(t, true, "", "/resource1", "/{resource}")
	testKeyMatch3Func(t, false, "", "/myid", "/{id}/using/{resId}")
	testKeyMatch3Func(t, true, "", "/myid/using/myresid", "/{id}/using/{resId}")

	testKeyMatch3Func(t, false, "", "/proxy/myid", "/proxy/{id}/*")
	testKeyMatch3Func(t, true, "", "/proxy/myid/", "/proxy/{id}/*")
	testKeyMatch3Func(t, true, "", "/proxy/myid/res", "/proxy/{id}/*")
	testKeyMatch3Func(t, true, "", "/proxy/myid/res/res2", "/proxy/{id}/*")
	testKeyMatch3Func(t, true, "", "/proxy/myid/res/res2/res3", "/proxy/{id}/*")
	testKeyMatch3Func(t, false, "", "/proxy/", "/proxy/{id}/*")
}

func TestKeyMatch4Func(t *testing.T) {
	testKeyMatch4Func(t, false, "keyMatch4: Expected 2 arguments, but got 1", "/parent/123/child/123")
	testKeyMatch4Func(t, false, "keyMatch4: Expected 2 arguments, but got 3", "/parent/123/child/123", "/parent/{id}/child/{id}", true)
	testKeyMatch4Func(t, false, "keyMatch4: Argument must be a string", "/parent/123/child/123", true)

	testKeyMatch4Func(t, true, "", "/parent/123/child/123", "/parent/{id}/child/{id}")
	testKeyMatch4Func(t, false, "", "/parent/123/child/456", "/parent/{id}/child/{id}")

	testKeyMatch4Func(t, true, "", "/parent/123/child/123", "/parent/{id}/child/{another_id}")
	testKeyMatch4Func(t, true, "", "/parent/123/child/456", "/parent/{id}/child/{another_id}")

}

func TestIPMatchFunc(t *testing.T) {
	testIPMatchFunc(t, false, "ipMatch: Expected 2 arguments, but got 1", "192.168.2.123")
	testIPMatchFunc(t, false, "ipMatch: Argument must be a string", "192.168.2.123", 128)
	testIPMatchFunc(t, true, "", "192.168.2.123", "192.168.2.0/24")
}

func TestGlobMatch(t *testing.T) {
	testGlobMatch(t, "/foo", "/foo", true)
	testGlobMatch(t, "/foo", "/foo*", true)
	testGlobMatch(t, "/foo", "/foo/*", false)
	testGlobMatch(t, "/foo/bar", "/foo", false)
	testGlobMatch(t, "/foo/bar", "/foo*", false)
	testGlobMatch(t, "/foo/bar", "/foo/*", true)
	testGlobMatch(t, "/foobar", "/foo", false)
	testGlobMatch(t, "/foobar", "/foo*", true)
	testGlobMatch(t, "/foobar", "/foo/*", false)

	testGlobMatch(t, "/foo", "*/foo", true)
	testGlobMatch(t, "/foo", "*/foo*", true)
	testGlobMatch(t, "/foo", "*/foo/*", false)
	testGlobMatch(t, "/foo/bar", "*/foo", false)
	testGlobMatch(t, "/foo/bar", "*/foo*", false)
	testGlobMatch(t, "/foo/bar", "*/foo/*", true)
	testGlobMatch(t, "/foobar", "*/foo", false)
	testGlobMatch(t, "/foobar", "*/foo*", true)
	testGlobMatch(t, "/foobar", "*/foo/*", false)

	testGlobMatch(t, "/prefix/foo", "*/foo", false)
	testGlobMatch(t, "/prefix/foo", "*/foo*", false)
	testGlobMatch(t, "/prefix/foo", "*/foo/*", false)
	testGlobMatch(t, "/prefix/foo/bar", "*/foo", false)
	testGlobMatch(t, "/prefix/foo/bar", "*/foo*", false)
	testGlobMatch(t, "/prefix/foo/bar", "*/foo/*", false)
	testGlobMatch(t, "/prefix/foobar", "*/foo", false)
	testGlobMatch(t, "/prefix/foobar", "*/foo*", false)
	testGlobMatch(t, "/prefix/foobar", "*/foo/*", false)

	testGlobMatch(t, "/prefix/subprefix/foo", "*/foo", false)
	testGlobMatch(t, "/prefix/subprefix/foo", "*/foo*", false)
	testGlobMatch(t, "/prefix/subprefix/foo", "*/foo/*", false)
	testGlobMatch(t, "/prefix/subprefix/foo/bar", "*/foo", false)
	testGlobMatch(t, "/prefix/subprefix/foo/bar", "*/foo*", false)
	testGlobMatch(t, "/prefix/subprefix/foo/bar", "*/foo/*", false)
	testGlobMatch(t, "/prefix/subprefix/foobar", "*/foo", false)
	testGlobMatch(t, "/prefix/subprefix/foobar", "*/foo*", false)
	testGlobMatch(t, "/prefix/subprefix/foobar", "*/foo/*", false)
}
