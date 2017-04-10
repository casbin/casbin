package main

func testBasicModel() {
	enforcer := &Enforcer{}
	enforcer.init("examples/basic_model.conf", "examples/basic_policy.txt")

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
	enforcer.init("examples/rbac_model.conf", "examples/basic_policy.txt")
}

func main() {
	// testBasicModel()
	testRBACModel()
}