package casbin

import (
	"github.com/hsluoyz/casbin/util"
	"log"
	"reflect"
	"testing"
)

func testEnforce(t *testing.T, e *Enforcer, sub string, obj string, act string, res bool) {
	if e.Enforce(sub, obj, act) != res {
		t.Errorf("%s, %s, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	}
}

func TestBasicModel(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/basic_model.conf", "examples/basic_policy.csv")

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
	e.Init("examples/basic_model_with_root.conf", "examples/basic_policy.csv")

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
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelWithResourceRoles(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model_with_resource_roles.conf", "examples/rbac_policy_with_resource_roles.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func getAttr(name string, attr string) string {
	// This is the same as:
	//
	// alice.domain = domain1
	// bob.domain = domain2
	// data1.domain = domain1
	// data2.domain = domain2

	if attr != "domain" {
		return "unknown"
	}

	if name == "alice" || name == "data1" {
		return "domain1"
	} else if name == "bob" || name == "data2" {
		return "domain2"
	} else {
		return "unknown"
	}
}

func getAttrFunc(args ...interface{}) (interface{}, error) {
	name := args[0].(string)
	attr := args[1].(string)

	return (string)(getAttr(name, attr)), nil
}

func TestABACModel(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/abac_model.conf", "")

	e.AddSubjectAttributeFunction(getAttrFunc)
	e.AddObjectAttributeFunction(getAttrFunc)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", true)
}

type testUser struct {
	name   string
	domain string
}

func newTestUser(name string, domain string) *testUser {
	u := testUser{}
	u.name = name
	u.domain = domain
	return &u
}

func (u *testUser) getAttribute(attributeName string) string {
	ru := reflect.ValueOf(u)
	f := reflect.Indirect(ru).FieldByName(attributeName)
	return f.String()
}

type testResource struct {
	name   string
	domain string
}

func newTestResource(name string, domain string) *testResource {
	r := testResource{}
	r.name = name
	r.domain = domain
	return &r
}

func (u *testResource) getAttribute(attributeName string) string {
	ru := reflect.ValueOf(u)
	f := reflect.Indirect(ru).FieldByName(attributeName)
	return f.String()
}

func TestABACModel2(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/abac_model.conf", "")

	alice := newTestUser("alice", "domain1")
	bob := newTestUser("bob", "domain2")
	data1 := newTestResource("data1", "domain1")
	data2 := newTestResource("data2", "domain2")

	log.Println(alice.getAttribute("domain"))
	log.Println(bob.getAttribute("domain"))
	log.Println(data1.getAttribute("domain"))
	log.Println(data2.getAttribute("domain"))
}

func TestKeymatchModel(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/keymatch_model.conf", "examples/keymatch_policy.csv")

	testEnforce(t, e, "alice", "/alice_data/resource1", "GET", true)
	testEnforce(t, e, "alice", "/alice_data/resource1", "POST", true)
	testEnforce(t, e, "alice", "/alice_data/resource2", "GET", true)
	testEnforce(t, e, "alice", "/alice_data/resource2", "POST", false)
	testEnforce(t, e, "alice", "/bob_data/resource1", "GET", false)
	testEnforce(t, e, "alice", "/bob_data/resource1", "POST", false)
	testEnforce(t, e, "alice", "/bob_data/resource2", "GET", false)
	testEnforce(t, e, "alice", "/bob_data/resource2", "POST", false)
	testEnforce(t, e, "bob", "/alice_data/resource1", "GET", false)
	testEnforce(t, e, "bob", "/alice_data/resource1", "POST", false)
	testEnforce(t, e, "bob", "/alice_data/resource2", "GET", true)
	testEnforce(t, e, "bob", "/alice_data/resource2", "POST", false)
	testEnforce(t, e, "bob", "/bob_data/resource1", "GET", false)
	testEnforce(t, e, "bob", "/bob_data/resource1", "POST", true)
	testEnforce(t, e, "bob", "/bob_data/resource2", "GET", false)
	testEnforce(t, e, "bob", "/bob_data/resource2", "POST", true)
}

func testGetRoles(t *testing.T, e *Enforcer, name string, res []string) {
	myRes := e.GetRoles(name)
	log.Print("Roles for ", name, ": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error("Roles for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetRoles(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})
	testGetRoles(t, e, "non_exist", []string{})
}

func testStringList(t *testing.T, title string, f func() []string, res []string) {
	myRes := f()
	log.Print(title+": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error(title+": ", myRes, ", supposed to be ", res)
	}
}

func TestGetList(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testStringList(t, "Subjects", e.GetAllSubjects, []string{"alice", "bob", "data2_admin"})
	testStringList(t, "Objeccts", e.GetAllObjects, []string{"data1", "data2"})
	testStringList(t, "Actions", e.GetAllActions, []string{"read", "write"})
	testStringList(t, "Roles", e.GetAllRoles, []string{"data2_admin"})
}

func testGetPolicy(t *testing.T, e *Enforcer, res [][]string) {
	myRes := e.GetPolicy()
	log.Print("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredPolicy(t *testing.T, e *Enforcer, fieldIndex int, fieldValue string, res [][]string) {
	myRes := e.GetFilteredPolicy(fieldIndex, fieldValue)
	log.Print("Policy for ", fieldValue, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy for ", fieldValue, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetGroupingPolicy(t *testing.T, e *Enforcer, res [][]string) {
	myRes := e.GetGroupingPolicy()
	log.Print("Grouping policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy: ", myRes, ", supposed to be ", res)
	}
}

func TestGetPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	testGetFilteredPolicy(t, e, 0, "alice", [][]string{{"alice", "data1", "read"}})
	testGetFilteredPolicy(t, e, 0, "bob", [][]string{{"bob", "data2", "write"}})
	testGetFilteredPolicy(t, e, 0, "data2_admin", [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetFilteredPolicy(t, e, 1, "data1", [][]string{{"alice", "data1", "read"}})
	testGetFilteredPolicy(t, e, 1, "data2", [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetFilteredPolicy(t, e, 2, "read", [][]string{{"alice", "data1", "read"}, {"data2_admin", "data2", "read"}})
	testGetFilteredPolicy(t, e, 2, "write", [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "write"}})

	testGetGroupingPolicy(t, e, [][]string{{"alice", "data2_admin"}})
}

func TestReloadPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func TestSavePolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.SavePolicy()
}

func TestModifyPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.RemovePolicy([]string{"alice", "data1", "read"})
	e.RemovePolicy([]string{"bob", "data2", "write"})
	e.RemovePolicy([]string{"alice", "data1", "read"})
	e.AddPolicy([]string{"eve", "data3", "read"})

	testGetPolicy(t, e, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"eve", "data3", "read"}})
}

func TestModifyGroupingPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.RemoveGroupingPolicy([]string{"alice", "data2_admin"})
	e.AddGroupingPolicy([]string{"bob", "data1_admin"})
	e.AddGroupingPolicy([]string{"eve", "data3_admin"})

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{"data1_admin"})
	testGetRoles(t, e, "eve", []string{"data3_admin"})
	testGetRoles(t, e, "non_exist", []string{})
}

func TestEnable(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/basic_model.conf", "examples/basic_policy.csv")

	e.Enable(false)
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", true)
	testEnforce(t, e, "bob", "data1", "write", true)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", true)

	e.Enable(true)
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func benchmarkEnforce(b *testing.B, e *Enforcer, sub string, obj string, act string, res bool) {
	if e.Enforce(sub, obj, act) != res {
		b.Errorf("%s, %s, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	}
}

func BenchmarkBasicModel(b *testing.B) {
	e := &Enforcer{}
	e.Init("examples/basic_model.conf", "examples/basic_policy.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkEnforce(b, e, "alice", "data1", "read", true)
	}
}

func BenchmarkRBACModel(b *testing.B) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkEnforce(b, e, "alice", "data2", "read", true)
	}
}
