package casbin

import "testing"

func TestDBSavePolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	a := newDbAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
	a.open()
	a.savePolicy(e.model)
	a.close()
}

func TestDBSaveAndLoadPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/rbac_model.conf", "examples/rbac_policy.csv")

	a := newDbAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
	a.open()
	a.savePolicy(e.model)

	clearPolicy(e.model)
	testGetPolicy(t, e, [][]string{})

	a.loadPolicy(e.model)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	a.close()
}
