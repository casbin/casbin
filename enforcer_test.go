package casbin

import (
	"testing"
	"log"
)

func testEnforce(t *testing.T, e *Enforcer, sub string, obj string, act string, res bool) {
	if e.enforce(sub, obj, act) != res {
		t.Errorf("%s, %s, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	}
}

func TestBasicModel(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/basic_model.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestBasicModelWithRoot(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/basic_model_with_root.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
	testEnforce(t, e, "root", "data1", "read", true)
	testEnforce(t, e, "root", "data1", "write", true)
	testEnforce(t, e, "root", "data2", "read", true)
	testEnforce(t, e, "root", "data2", "write", true)
}

func TestRBACModel(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func testKeyMatch(t *testing.T, e *Enforcer, key1 string, key2 string, res bool) {
	myRes := e.keyMatch(key1, key2)
	log.Printf("%s < %s: %t", key1, key2, myRes)

	if myRes != res {
		t.Errorf("%s < %s: %t, supposed to be %t", key1, key2, !res, res)
	}
}

func TestKeyMatch(t *testing.T) {
	e := &Enforcer{}

	testKeyMatch(t, e, "/foo", "/foo", true)
	testKeyMatch(t, e, "/foo", "/foo*", true)
	testKeyMatch(t, e, "/foo", "/foo/*", false)
	testKeyMatch(t, e, "/foo/bar", "/foo", false)
	testKeyMatch(t, e, "/foo/bar", "/foo*", true)
	testKeyMatch(t, e, "/foo/bar", "/foo/*", true)
	testKeyMatch(t, e, "/foobar", "/foo", false)
	testKeyMatch(t, e, "/foobar", "/foo*", true)
	testKeyMatch(t, e, "/foobar", "/foo/*", false)
}

func arrayEquals(a []string, b []string) bool {
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

func array2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !arrayEquals(v, b[i]) {
			return false
		}
	}
	return true
}

func testGetRoles(t *testing.T, e *Enforcer, name string, res []string) {
	myRes := e.getRoles(name)
	log.Print("Roles for ", name, ": ", myRes)

	if !arrayEquals(res, myRes) {
		t.Error("Roles for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetRoles(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})
	testGetRoles(t, e, "non_exist", []string{})
}

func testGetPolicy(t *testing.T, e *Enforcer, res [][]string) {
	myRes := e.getPolicy()
	log.Print("Policy: ", myRes)

	if !array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredPolicy(t *testing.T, e *Enforcer, fieldIndex int, fieldValue string, res [][]string) {
	myRes := e.getFilteredPolicy(fieldIndex, fieldValue)
	log.Print("Policy for ", fieldValue, ": ", myRes)

	if !array2DEquals(res, myRes) {
		t.Error("Policy for ", fieldValue, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetPolicy(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetFilteredPolicy(t, e, 0, "alice", [][]string{{"alice", "data1", "read"}})
	testGetFilteredPolicy(t, e, 0, "bob", [][]string{{"bob", "data2", "write"}})
	testGetFilteredPolicy(t, e, 0, "data2_admin", [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetFilteredPolicy(t, e, 1, "data1", [][]string{{"alice", "data1", "read"}})
	testGetFilteredPolicy(t, e, 1, "data2", [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetFilteredPolicy(t, e, 2, "read", [][]string{{"alice", "data1", "read"}, {"data2_admin", "data2", "read"}})
	testGetFilteredPolicy(t, e, 2, "write", [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "write"}})
}
