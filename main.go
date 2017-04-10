package main

import "fmt"

func testBasicModel() {
	enforcer := &Enforcer{}
	enforcer.init("examples/basic_model.conf", "examples/basic_policy.csv")

	enforcer.enforce("alice", "data1", "read")
	enforcer.enforce("alice", "data1", "write")
	enforcer.enforce("alice", "data2", "read")
	enforcer.enforce("alice", "data2", "write")
	enforcer.enforce("bob", "data1", "read")
	enforcer.enforce("bob", "data1", "write")
	enforcer.enforce("bob", "data2", "read")
	enforcer.enforce("bob", "data2", "write")
}

func testRBACModel() {
	enforcer := &Enforcer{}
	enforcer.init("examples/rbac_model.conf", "examples/rbac_policy.csv")
}

func testRole() {
	rm := newRoleManager(3)
	rm.addLink("u1", "g1")
	rm.addLink("u2", "g1")
	rm.addLink("u3", "g2")
	rm.addLink("g1", "g3")

	for _, u := range []string{"u1", "u2", "u3"} {
		for _, g := range []string{"g1", "g2", "g3"} {
			res := rm.hasLink(u, g)
			fmt.Print(u + ", " + g + ": ")
			fmt.Println(res)
		}
	}
}

func main() {
	// testBasicModel()
	testRBACModel()
	// testRole()
}