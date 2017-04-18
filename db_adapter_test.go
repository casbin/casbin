package casbin

import "testing"

func TestDBLoadPolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/basic_model.conf", "")

	a := newDbAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
	a.open()
	a.loadPolicy(e.model)
	printPolicy(e.model)
	a.close()
}

func TestDBSavePolicy(t *testing.T) {
	e := &Enforcer{}
	e.Init("examples/basic_model.conf", "examples/basic_policy.csv")

	a := newDbAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
	a.open()
	a.savePolicy(e.model)
	a.close()
}
