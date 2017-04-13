casbin
====

[![Go Report Card](https://goreportcard.com/badge/github.com/hsluoyz/casbin)](https://goreportcard.com/report/github.com/hsluoyz/casbin)
[![Build Status](https://travis-ci.org/hsluoyz/casbin.svg?branch=master)](https://travis-ci.org/hsluoyz/casbin)
[![Godoc](https://godoc.org/github.com/hsluoyz/casbin?status.png)](https://godoc.org/github.com/hsluoyz/casbin)


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
go get github.com/hsluoyz/casbin/...
```

Get started
--

1. Initialize an enforcer by specifying a model CONF file and the policy file.

```golang
e := &Enforcer{}
e.init("examples/basic_model.conf", "examples/basic_policy.csv")
```

2. Add the enforcement hook into your code before the access happens.

```golang
sub := "alice"
obj := "data1"
act := "read"

if e.enforce(sub, obj, act) == true {
    // permit alice to read data1
} else {
    // deny the request, show an error
}
```

3. You can get the roles for a user with our management API.

```golang
roles := e.getRoles("alice")
```

4. Please refer to the ``_test.go`` files for more usage.

Credits
--

- [github.com/lxmgo/config](https://github.com/lxmgo/config)
- [github.com/Knetic/govaluate](https://github.com/Knetic/govaluate)

License
--

This project is licensed under the Apache 2.0 license.
