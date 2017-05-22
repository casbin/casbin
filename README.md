casbin
====

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin)](https://goreportcard.com/report/github.com/casbin/casbin)
[![Build Status](https://travis-ci.org/casbin/casbin.svg?branch=master)](https://travis-ci.org/casbin/casbin)
[![Coverage Status](https://coveralls.io/repos/github/casbin/casbin/badge.svg?branch=master)](https://coveralls.io/github/casbin/casbin?branch=master)
[![Godoc](https://godoc.org/github.com/casbin/casbin?status.svg)](https://godoc.org/github.com/casbin/casbin)
[![Release](https://img.shields.io/github/release/casbin/casbin.svg)](https://github.com/casbin/casbin/releases/latest)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/casbin/lobby)

**Note**: The plugins and middleware based on casbin can be found at: https://github.com/casbin

![casbin Logo](casbin-logo.png)

casbin is a powerful and efficient open-source access control library for Golang projects. It provides support for enforcing authorization based on various models. By far, the access control models supported by casbin are:

1. [**ACL (Access Control List)**](https://en.wikipedia.org/wiki/Access_control_list)
2. **ACL with [superuser](https://en.wikipedia.org/wiki/Superuser)**
3. **ACL without users**: especially useful for systems that don't have authentication or user log-ins.
3. **ACL without resources**: some scenarios may target for a type of resources instead of an individual resource by using permissions like ``write-article``, ``read-log``. It doesn't control the access to a specific article or log.
4. **[RBAC (Role-Based Access Control)](https://en.wikipedia.org/wiki/Role-based_access_control)**
5. **RBAC with resource roles**: both users and resources can have roles (or groups) at the same time.
6. **RBAC with domains/tenants**: users can have different role sets for different domains/tenants.
7. **[ABAC (Attribute-Based Access Control)](https://en.wikipedia.org/wiki/Attribute-Based_Access_Control)**
8. **[RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer)**

In casbin, an access control model is abstracted into a CONF file based on the **PERM metamodel (Policy, Effect, Request, Matchers)**. So switching or upgrading the authorization mechanism for a project is just as simple as modifying a configuration. You can customize your own access control model by combining the available models. For example, you can get RBAC roles and ABAC attributes together inside one model and share one set of policy rules.

The most basic and simplest model in casbin is ACL. ACL's model CONF is:

```ini
# Request definition
[request_definition]
r = sub, obj, act

# Policy definition
[policy_definition]
p = sub, obj, act

# Policy effect
[policy_effect]
e = some(where (p.eft == allow))

# Matchers
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

An example policy for ACL model is like:

```
p, alice, data1, read
p, bob, data2, write
```

It means:

- alice can read data1
- bob can write data2

## Features

What casbin does:

1. enforce the policy in the classic ``{subject, object, action}`` form or a customized form as you defined.
2. handle the storage of the access control model and its policy.
3. manage the role-user mappings and role-role mappings (aka role hierarchy in RBAC).
4. support built-in superuser like ``root`` or ``administrator``. A superuser can do anything without explict permissions.
5. multiple built-in operators to support the rule matching. For example, ``keyMatch`` can map a resource key ``/foo/bar`` to the pattern ``/foo*``.

What casbin does NOT do:

1. authentication (aka verify ``username`` and ``password`` when a user logs in)
2. manage the list of users or roles. I believe it's more convenient for the project itself to manage these entities. Users usually have their passwords, and casbin is not designed as a password container. However, casbin stores the user-role mapping for the RBAC scenario. 

## Installation

```
go get github.com/casbin/casbin
```

## Get started

1. Customize the casbin config file ``casbin.conf`` to your need. Its default content is:

```ini
[default]
# The file path to the model:
model_path = examples/basic_model.conf

# The persistent method for policy, can be two values: file or database.
# policy_backend = file
# policy_backend = database
policy_backend = file

[file]
# The file path to the policy:
policy_path = examples/basic_policy.csv

[database]
driver = mysql
data_source = root:@tcp(127.0.0.1:3306)/
```

It means uses ``basic_model.conf`` as the model and ``basic_policy.csv`` as the policy.

2. Initialize an enforcer by specifying the path to the casbin configuration file:

```go
e := casbin.NewEnforcer("path/to/casbin.conf")
```

Note: you can also initialize an enforcer directly with a file path or database, see ``Persistence`` section for details.

3. Add an enforcement hook into your code right before the access happens:

```go
sub := "alice" // the user that wants to access a resource.
obj := "data1" // the resource that is going to be accessed.
act := "read" // the operation that the user performs on the resource.

if e.Enforce(sub, obj, act) == true {
    // permit alice to read data1
} else {
    // deny the request, show an error
}
```

4. Besides the static policy file, casbin also provides API for permission management at run-time. For example, You can get all the roles assigned to a user as below:

```go
roles := e.GetRoles("alice")
```

5. Please refer to the ``_test.go`` files for more usage.

## Syntax for models

A model CONF should have at least four sections: ``[request_definition], [policy_definition], [policy_effect], [matchers]``. If the model uses RBAC, it should also add ``[role_definition]``. The comments start with ``#``.

### Request definition

``[request_definition]`` is the definition for the access request. It defines the arguments in ``e.Enforce(...)`` function.

```ini
[request_definition]
r = sub, obj, act
```

``sub, obj, act`` represents the classic triple: accessing entity (Subject), accessed resource (Object) and the access method (Action). However, you can customize your own request form, like ``sub, act`` if you don't need to specify an particular resource, or ``sub, sub2, obj, act`` if you somehow have two accessing entities.

### Policy definition

``[policy_definition]`` is the definition for the policy. It defines the meaning of the policy. For example, we have the following model:

```ini
[policy_definition]
p = sub, obj, act
p2 = sub, act
```

And we have the following policy (if in a policy file)

```
p, alice, data1, read
p2, bob, write-all-objects
```

Each line in a policy is called a policy rule. Each policy rule starts with a ``policy type``, e.g., `p`, `p2`. It is used to match the policy definition if there are multiple definitions. The above policy shows this mapping:

(alice, data1, read) -> (p.sub, p.obj, p.act)
(bob, write-all-objects) -> (p2.sub, p2.act)

For common cases, the user doesn't have multiple policy definitions, so probably you will only use policy type ``p``.

### Policy effect

``[policy_effect]`` is the definition for the policy effect. It defines whether the access request should be approved if multiple policy rules match the request. For example, one rule permits and the other denies.

```ini
[policy_effect]
e = some(where (p.eft == allow))
```
e = !any(where (p.eft == deny))
The above policy effect means if there's any matched policy rule of ``allow``, the final effect is ``allow`` (aka allow-override). ``p.eft`` is the effect for a policy, it can be ``allow`` or ``deny``. It's optional and the default value is ``allow``. So as we didn't specify it above, it uses the default value.

Another example for policy effect is:

```ini
e = !any(where (p.eft == deny))
```

It means if there should be no matched policy rules of``deny`` (aka deny-override). ``some`` means: if there exists one matched policy rule. ``any`` means: all matched policy rules. The policy effect can even be connected with logic expressions:

```ini
e = some(where (p.eft == allow)) && !any(where (p.eft == deny))
```

It means at least one matched policy rule of``allow``, and there should be matched policy rules of``deny``.

### Matchers

``[matchers]`` is the definition for policy matchers. The matchers are expressions. It defines how the policy rules are evaluated against the request.

```ini
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

The above matcher is the simplest, it means that the subject, object and action in a request should match the ones in a policy rule.

You can use arithmetic like ``+, -, *, /`` and logical operators like ``&&, ||, !`` in matchers.

#### Functions in matchers

You can even specify functions in a matcher. You can use the built-in functions or specify your own function. The supported built-in functions are:

- ``keyMatch(arg1, arg2)``: arg1 and arg2 are usually paths or URLs. arg2 can have pattern (*). It returns whether arg1 matches arg2.
- ``regexMatch(arg1, arg2)``: arg1 can be any string. arg2 is a regular expression. It returns whether arg1 matches arg2.

Please refer to [keymatch_model.conf](https://github.com/casbin/casbin/blob/master/examples/keymatch_model.conf) for examples.

#### How to add a customized function

First prepare your function. It takes several parameters and return a bool:

```go
func KeyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}
```

Then wrap it with ``interface{}`` types:

```go
func KeyMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return (bool)(KeyMatch(name1, name2)), nil
}
```

At last, register the function to the casbin enforcer:

```go
e.AddFunction("my_func", KeyMatchFunc)
```

Now, you can use the function in your model CONF like this:

```ini
[matchers]
m = r.sub == p.sub && my_func(r.obj, p.obj) && r.act == p.act
```

### Role definition (optional)

``[role_definition]`` is the definition for the RBAC role inheritance relations. Casbin supports multiple instances of RBAC systems, e.g., users can have roles and their inheritance relations, and resources can have roles and their inheritance relations too. These two RBAC systems won't interfere.

This section is optional. If you don't use RBAC roles in the model, then omit this section.

```ini
[role_definition]
g = _, _
g2 = _, _
```

The above role definition shows that ``g`` is a RBAC system, and ``g2`` is another RBAC system. ``_, _`` means there are two parties inside an inheritance relation. As a common case, you usually use ``g`` alone if you only need roles on users. and you can use ``g`` and ``g2`` when you need roles (or groups) on both users and resources. Please see the [rbac_model.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model.conf) and [rbac_model_with_resource_roles.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model_with_resource_roles.conf) for examples.

Casbin stores the actual user-role mapping (or resource-role mapping if you are using roles on resources) in the policy, for example:

```
p, data2_admin, data2, read
g, alice, data2_admin
```

It means ``alice`` inherits/is a member of role ``data2_admin``. ``alice`` here can be a user, a resource or a role. Casbin only recognizes it as a string.

Then in a matcher, you should check the role as below:

```ini
[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

It means ``sub`` in the request should has the role ``sub`` in the policy.

There are several things to note:

1. Casbin only stores the user-role mapping.
2. Casbin doesn't verify whether a user is a valid user, or role is a valid role. That should be taken care of by authentication.
3. Do not use the same name for a user and a role inside a RBAC system, because Casbin recognizes users and roles as strings, and there's no way for Casbin to know whether you are specifying user ``alice`` or role ``alice``. You can simply solve it by using ``role_alice``.
4. If ``A`` has role ``B``, ``B`` has role ``C``, then ``A`` has role ``C``. This transitivity is infinite for now.

### Role definition with domains/tenants (optional)

The RBAC roles in Casbin can be global or domain-specific. Domain-specify roles mean that the roles for a user can be different when the user is at different domains/tenants. This is very useful for large systems like a cloud, as the users are usually in different tenants.

The role definition with domains/tenants should be something like:

```ini
[role_definition]
g = _, _, _
```

The 3rd ``_`` means the name of domain/tenant, this part should not be changed. Then the policy can be:

```
p, admin, tenant1, data1, read
p, admin, tenant2, data2, read

g, alice, admin, tenant1
g, alice, user, tenant2
```

It means ``admin`` role in ``tenant1`` can read ``data1``. And ``alice`` has ``admin`` role in ``tenant1``, and has ``user`` role in ``tenant2``. So she can read ``data1``. However, since ``alice`` is not an ``admin`` in ``tenant2``, she cannot read ``data2``.

Then in a matcher, you should check the role as below:

```ini
[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
```

Please see the [rbac_model_with_domains.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model_with_domains.conf) for examples.

## Persistence

The model and policy can be persisted in casbin with the following restrictions:

Persist Method | casbin Model | casbin Policy | Usage
----|------|----|----
File | Load only | Load/Save | [Details](https://github.com/casbin/casbin#file)
Database (tested with [MySQL](https://www.mysql.com)) | Not supported | Load/Save | [Details](https://github.com/casbin/casbin#database)
[Cassandra](http://cassandra.apache.org) (NoSQL) | Not supported | Load/Save | [Details](https://github.com/casbin/cassandra_adapter)

We think the model represents the access control model that our customer uses and is not often modified at run-time, so we don't implement an API to modify the current model or save the model into a file. And the model cannot be loaded from or saved into a database. The model file should be in .CONF format.

The policy is much more dynamic than model and can be loaded from a file/database or saved to a file/database at any time. As for file persistence, the policy file should be in .CSV (Comma-Separated Values) format. As for the database backend, casbin should support all relational DBMSs but I only tested with MySQL. casbin has no built-in database with it, you have to setup a database on your own. Let me know if there are any compatibility issues here. casbin will automatically create a database named ``casbin`` and use it for policy storage. So make sure your provided credential has the related privileges for the database you use.

### File

Below shows how to initialize an enforcer from file:

```go
// Initialize an enforcer with a model file and a policy file.
e := casbin.NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
```

### Database

Below shows how to initialize an enforcer from database. it connects to a MySQL DB on 127.0.0.1:3306 with root and blank password.

```go
// Initialize an enforcer with a model file and policy from database.
e := casbin.NewEnforcer("examples/basic_model.conf", "mysql", "root:@tcp(127.0.0.1:3306)/")
```

### Use your own storage adapter

In casbin, both the above file and database storage is implemented as an adapter. You can use your own adapter like below:

```go
// Initialize an enforcer with an adapter.
adapter := persist.NewFileAdapter("examples/basic_policy.csv") // or replace with your own adapter.
e := casbin.NewEnforcer("examples/basic_model.conf", adapter)
```

An adapter should implement two methods:``LoadPolicy(model model.Model)`` and ``SavePolicy(model model.Model)``. To keep light-weight, we don't put all adapters' code in this main library. You can choose officially supported adapters from: https://github.com/casbin and use it like a plugin as above.

### Load/Save at run-time

You may also want to reload the model, reload the policy or save the policy after initialization:

```go
// Reload the model from the model CONF file.
e.LoadModel()

// Reload the policy from file/database.
e.LoadPolicy()

// Save the current policy (usually after changed with casbin API) back to file/database.
e.SavePolicy()
```

## Examples

Model | Model file | Policy file
----|------|----
ACL | [basic_model.conf](https://github.com/casbin/casbin/blob/master/examples/basic_model.conf) | [basic_policy.csv](https://github.com/casbin/casbin/blob/master/examples/basic_policy.csv)
ACL with superuser | [basic_model_with_root.conf](https://github.com/casbin/casbin/blob/master/examples/basic_model_with_root.conf) | [basic_policy.csv](https://github.com/casbin/casbin/blob/master/examples/basic_policy.csv)
ACL without users | [basic_model_without_users.conf](https://github.com/casbin/casbin/blob/master/examples/basic_model_without_users.conf) | [basic_policy_without_users.csv](https://github.com/casbin/casbin/blob/master/examples/basic_policy_without_users.csv)
ACL without resources | [basic_model_without_resources.conf](https://github.com/casbin/casbin/blob/master/examples/basic_model_without_resources.conf) | [basic_policy_without_resources.csv](https://github.com/casbin/casbin/blob/master/examples/basic_policy_without_resources.csv)
RBAC | [rbac_model.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model.conf)  | [rbac_policy.csv](https://github.com/casbin/casbin/blob/master/examples/rbac_policy.csv)
RBAC with resource roles | [rbac_model_with_resource_roles.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model_with_resource_roles.conf)  | [rbac_policy_with_resource_roles.csv](https://github.com/casbin/casbin/blob/master/examples/rbac_policy_with_resource_roles.csv)
RBAC with domains/tenants | [rbac_model_with_domains.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model_with_domains.conf)  | [rbac_policy_with_domains.csv](https://github.com/casbin/casbin/blob/master/examples/rbac_policy_with_domains.csv)
ABAC | [abac_model.conf](https://github.com/casbin/casbin/blob/master/examples/abac_model.conf)  | N/A
RESTful | [keymatch_model.conf](https://github.com/casbin/casbin/blob/master/examples/keymatch_model.conf)  | [keymatch_policy.csv](https://github.com/casbin/casbin/blob/master/examples/keymatch_policy.csv)

## Our users

- [Beego](https://github.com/astaxie/beego): An open-source, high-performance web framework for Go, see details in: [plugins/authz/authz_test.go](https://github.com/astaxie/beego/blob/master/plugins/authz/authz_test.go)
- [Docker](https://github.com/docker/docker): The world's leading software container platform, via plugin: [casbin-authz-plugin](https://github.com/casbin/casbin-authz-plugin)
- [pybbs-go](https://github.com/tomoya92/pybbs-go): A simple BBS with fine-grained permission management based on [Beego](https://github.com/astaxie/beego)
- [Tango](https://github.com/lunny/tango): Micro & pluggable web framework for Go, via plugin: [authz](https://github.com/tango-contrib/authz)
- [chi](https://github.com/pressly/chi): A lightweight, idiomatic and composable router for building HTTP services, via plugin: [chi-authz](https://github.com/casbin/chi-authz)

## License

This project is licensed under the [Apache 2.0 license](https://github.com/casbin/casbin/blob/master/LICENSE).

## Contact

If you have any issues or feature requests, please contact us. PR is welcomed.
- https://github.com/casbin/casbin/issues
- hsluoyz@gmail.com
