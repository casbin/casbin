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
	my_res := e.keyMatch(key1, key2)
	log.Printf("%s, %s: %t", key1, key2, my_res)

	if my_res != res {
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

func TestGetRoles(t *testing.T) {
	e := &Enforcer{}
	e.init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	log.Print("Roles for alice: ", e.getRoles("alice"))
	log.Print("Roles for bob: ", e.getRoles("bob"))
}
