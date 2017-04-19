casbin
====

[![Go Report Card](https://goreportcard.com/badge/github.com/hsluoyz/casbin)](https://goreportcard.com/report/github.com/hsluoyz/casbin)
[![Build Status](https://travis-ci.org/hsluoyz/casbin.svg?branch=master)](https://travis-ci.org/hsluoyz/casbin)
[![Coverage Status](https://coveralls.io/repos/github/hsluoyz/casbin/badge.svg?branch=master)](https://coveralls.io/github/hsluoyz/casbin?branch=master)
[![Godoc](https://godoc.org/github.com/hsluoyz/casbin?status.svg)](https://godoc.org/github.com/hsluoyz/casbin)
[![Release](https://img.shields.io/github/release/hsluoyz/casbin.svg)](https://github.com/hsluoyz/casbin/releases/latest)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/casbin/lobby)

casbin is a powerful and efficient open-source access control library for Golang projects. It provides support for enforcing authorization based on various models like ACL, RBAC, ABAC.

In casbin, an access control model is abstracted into a CONF file based on the PERM metamodel (Policy, Effect, Request, Matchers). So switching or upgrading the authorization mechanism for a project is just as simple as modifying a configuration. A model CONF can be as simple as:

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

A simple policy for this model is a CSV like:

```csv
p, alice, data1, read
p, bob, data2, write
```

Features
--

What casbin does:

1. enforce the policy in the classic ``{subject, object, action}`` form or a customized form as you defined.
2. handle the storage of the access control model and its policy.
3. manage the role-user mappings and role-role mappings (aka role hierarchy in RBAC).
4. support built-in superuser like ``root`` or ``administrator``. A superuser can do anything without explict permissions.
5. multiple built-in operators to support the rule matching. For example, ``keyMatch`` can map a resource key ``/foo/bar`` to the pattern ``/foo*``.

What casbin does NOT do:

1. authentication (aka verify ``username`` and ``password`` when a user logs in)
2. manage the list of users or roles. I believe it's more convenient for the project itself to manage these entities. Users usually have their passwords, and casbin is not designed as a password container. However, casbin stores the user-role mapping for the RBAC scenario. 

Installation
--

```
go get github.com/hsluoyz/casbin
```

Get started
--

1. Initialize an enforcer by specifying a model CONF file and the policy file.

```golang
e := &Enforcer{}
e.Init("examples/basic_model.conf", "examples/basic_policy.csv")
```

2. Add the enforcement hook into your code before the access happens.

```golang
sub := "alice"
obj := "data1"
act := "read"

if e.Enforce(sub, obj, act) == true {
    // permit alice to read data1
} else {
    // deny the request, show an error
}
```

3. You can get the roles for a user with our management API.

```golang
roles := e.GetRoles("alice")
```

4. Please refer to the ``_test.go`` files for more usage.

Persistence
--

Both model and policy can be persisted in casbin with the following restrictions:

Persist Method | casbin Model | casbin Policy
----|------|----
File | Load only | Load/Save
Database (RDBMS) | Not supported | Load/Save

We think the model represents the access control model that our customer uses and is not often modified at run-time, so we don't implement an API to modify the current model or save the model into a file. And the model cannot be loaded from or saved into a database. The model file should be in .CONF format.

The policy is much more dynamic than model and can be loaded from a file/database or saved to a file/database at any time. As for file persistence, the policy file should be in .CSV (Comma-Separated Values) format. As for the database backend, casbin should support all relational DBMSs but I only tested with MySQL. casbin has no built-in database with it, you have to setup a database on your own. Let me know if there are any compatibility issues here. casbin will automatically create a database named ``casbin`` and use it for policy storage. So make sure your provided credential has the related privileges for the database you use.


Below shows how to initialize an enforcer from file:

```golang
e := &Enforcer{}
// Initialize an enforcer with a model file and a policy file.
e.InitWithFile("examples/basic_model.conf", "examples/basic_policy.csv")
```

Below shows how to initialize an enforcer from database. it connects to a MySQL DB on 127.0.0.1:3306 with root and blank password.

```golang
e := &Enforcer{}
// Initialize an enforcer with a model file and a policy from database.
e.InitWithDB("examples/basic_model.conf", "mysql", "root:@tcp(127.0.0.1:3306)/")
```

You may also want to reload the model, reload the policy or save the policy after initialization:

```golang
// Reload the model from the model CONF file.
e.LoadModel()

// Reload the policy from file/database.
e.LoadPolicy()

// Save the current policy (usually after changed with casbin API) back to file/database.
e.SavePolicy()
```

Examples
--

Model | Model file | Policy file
----|------|----
basic | [basic_model.conf](https://github.com/hsluoyz/casbin/blob/master/examples/basic_model.conf) | [basic_policy.csv](https://github.com/hsluoyz/casbin/blob/master/examples/basic_policy.csv)
basic with root | [basic_model_with_root.conf](https://github.com/hsluoyz/casbin/blob/master/examples/basic_model_with_root.conf) | [basic_policy.csv](https://github.com/hsluoyz/casbin/blob/master/examples/basic_policy.csv)
RESTful | [keymatch_model.conf](https://github.com/hsluoyz/casbin/blob/master/examples/keymatch_model.conf)  | [keymatch_policy.csv](https://github.com/hsluoyz/casbin/blob/master/examples/keymatch_policy.csv)
RBAC | [rbac_model.conf](https://github.com/hsluoyz/casbin/blob/master/examples/rbac_model.conf)  | [rbac_policy.csv](https://github.com/hsluoyz/casbin/blob/master/examples/rbac_policy.csv)
RBAC with resource roles | [rbac_model_with_resource_roles.conf](https://github.com/hsluoyz/casbin/blob/master/examples/rbac_model_with_resource_roles.conf)  | [rbac_policy_with_resource_roles.csv](https://github.com/hsluoyz/casbin/blob/master/examples/rbac_policy_with_resource_roles.csv)
ABAC | [abac_model.conf](https://github.com/hsluoyz/casbin/blob/master/examples/abac_model.conf)  | N/A

Credits
--

- [github.com/lxmgo/config](https://github.com/lxmgo/config)
- [github.com/Knetic/govaluate](https://github.com/Knetic/govaluate)

License
--

This project is licensed under the Apache 2.0 license.
