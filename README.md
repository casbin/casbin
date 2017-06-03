Casbin
====

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin)](https://goreportcard.com/report/github.com/casbin/casbin)
[![Build Status](https://travis-ci.org/casbin/casbin.svg?branch=master)](https://travis-ci.org/casbin/casbin)
[![Coverage Status](https://coveralls.io/repos/github/casbin/casbin/badge.svg?branch=master)](https://coveralls.io/github/casbin/casbin?branch=master)
[![Godoc](https://godoc.org/github.com/casbin/casbin?status.svg)](https://godoc.org/github.com/casbin/casbin)
[![Release](https://img.shields.io/github/release/casbin/casbin.svg)](https://github.com/casbin/casbin/releases/latest)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/casbin/lobby)
[![Sourcegraph Badge](https://sourcegraph.com/github.com/casbin/casbin/-/badge.svg)](https://sourcegraph.com/github.com/casbin/casbin?badge)

**Note**: The plugins and middleware based on Casbin can be found at: https://github.com/casbin

![casbin Logo](casbin-logo.png)

Casbin is a powerful and efficient open-source access control library for Golang projects. It provides support for enforcing authorization based on various models. By far, the access control models supported by Casbin are:

1. [**ACL (Access Control List)**](https://en.wikipedia.org/wiki/Access_control_list)
2. **ACL with [superuser](https://en.wikipedia.org/wiki/Superuser)**
3. **ACL without users**: especially useful for systems that don't have authentication or user log-ins.
3. **ACL without resources**: some scenarios may target for a type of resources instead of an individual resource by using permissions like ``write-article``, ``read-log``. It doesn't control the access to a specific article or log.
4. **[RBAC (Role-Based Access Control)](https://en.wikipedia.org/wiki/Role-based_access_control)**
5. **RBAC with resource roles**: both users and resources can have roles (or groups) at the same time.
6. **RBAC with domains/tenants**: users can have different role sets for different domains/tenants.
7. **[ABAC (Attribute-Based Access Control)](https://en.wikipedia.org/wiki/Attribute-Based_Access_Control)**
8. **[RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer)**
9. **Deny-override**: both allow and deny authorizations are supported, deny overrides the allow.

In Casbin, an access control model is abstracted into a CONF file based on the **PERM metamodel (Policy, Effect, Request, Matchers)**. So switching or upgrading the authorization mechanism for a project is just as simple as modifying a configuration. You can customize your own access control model by combining the available models. For example, you can get RBAC roles and ABAC attributes together inside one model and share one set of policy rules.

The most basic and simplest model in Casbin is ACL. ACL's model CONF is:

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

What Casbin does:

1. enforce the policy in the classic ``{subject, object, action}`` form or a customized form as you defined, both allow and deny authorizations are supported.
2. handle the storage of the access control model and its policy.
3. manage the role-user mappings and role-role mappings (aka role hierarchy in RBAC).
4. support built-in superuser like ``root`` or ``administrator``. A superuser can do anything without explict permissions.
5. multiple built-in operators to support the rule matching. For example, ``keyMatch`` can map a resource key ``/foo/bar`` to the pattern ``/foo*``.

What Casbin does NOT do:

1. authentication (aka verify ``username`` and ``password`` when a user logs in)
2. manage the list of users or roles. I believe it's more convenient for the project itself to manage these entities. Users usually have their passwords, and Casbin is not designed as a password container. However, Casbin stores the user-role mapping for the RBAC scenario. 

## Installation

```
go get github.com/casbin/casbin
```

## Get started

1. Customize the Casbin config file ``casbin.conf`` to your need. Its default content is:

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

2. Initialize an enforcer by specifying the path to the Casbin configuration file:

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

4. Besides the static policy file, Casbin also provides API for permission management at run-time. For example, You can get all the roles assigned to a user as below:

```go
roles := e.GetRoles("alice")
```

5. Please refer to the ``_test.go`` files for more usage.

## Syntax for models

See: [Model.md](https://github.com/casbin/casbin/blob/master/Model.md)

## Persistence

The model and policy can be persisted in Casbin with the following restrictions:

Persist Method | Casbin Model | Casbin Policy | Usage
----|------|----|----
File | Load only | Load/Save | [Details](https://github.com/casbin/casbin#file)
Database (tested with [MySQL](https://www.mysql.com)) | Not supported | Load/Save | [Details](https://github.com/casbin/casbin#database)
[Cassandra](http://cassandra.apache.org) (NoSQL) | Not supported | Load/Save | [Details](https://github.com/casbin/cassandra_adapter)

We think the model represents the access control model that our customer uses and is not often modified at run-time, so we don't implement an API to modify the current model or save the model into a file. And the model cannot be loaded from or saved into a database. The model file should be in .CONF format.

The policy is much more dynamic than model and can be loaded from a file/database or saved to a file/database at any time. As for file persistence, the policy file should be in .CSV (Comma-Separated Values) format. As for the database backend, Casbin should support all relational DBMSs but I only tested with MySQL. Casbin has no built-in database with it, you have to setup a database on your own. Let me know if there are any compatibility issues here. Casbin will automatically create a database named ``casbin`` and use it for policy storage. So make sure your provided credential has the related privileges for the database you use.

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

In Casbin, both the above file and database storage is implemented as an adapter. You can use your own adapter like below:

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

// Save the current policy (usually after changed with Casbin API) back to file/database.
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
Deny-override | [rbac_model_with_deny.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model_with_deny.conf)  | [rbac_policy_with_deny.csv](https://github.com/casbin/casbin/blob/master/examples/rbac_policy_with_deny.csv)

## Our users

### Web servers

- [Beego](https://github.com/astaxie/beego): An open-source, high-performance web framework for Go, via built-in plugin: [plugins/authz](https://github.com/astaxie/beego/blob/master/plugins/authz)
- [Caddy](https://github.com/mholt/caddy): Fast, cross-platform HTTP/2 web server with automatic HTTPS, via plugin: [caddy-authz](https://github.com/casbin/caddy-authz)
- [Gin](https://github.com/gin-gonic/gin): A HTTP web framework featuring a Martini-like API with much better performance, via plugin: [authz](https://github.com/gin-contrib/authz)
- [Revel](https://github.com/revel/revel): A high productivity, full-stack web framework for the Go language, via plugin: [revel-authz](https://github.com/casbin/revel-authz)
- [Echo](https://github.com/labstack/echo): High performance, minimalist Go web framework, via plugin: [echo-authz](https://github.com/labstack/echo-contrib/tree/master/casbin) (thanks to [@xqbumu](https://github.com/xqbumu))
- [Tango](https://github.com/lunny/tango): Micro & pluggable web framework for Go, via plugin: [authz](https://github.com/tango-contrib/authz)
- [Chi](https://github.com/pressly/chi): A lightweight, idiomatic and composable router for building HTTP services, via plugin: [chi-authz](https://github.com/casbin/chi-authz)

### Others

- [Docker](https://github.com/docker/docker): The world's leading software container platform, via plugin: [casbin-authz-plugin](https://github.com/casbin/casbin-authz-plugin)
- [pybbs-go](https://github.com/tomoya92/pybbs-go): A simple BBS with fine-grained permission management based on [Beego](https://github.com/astaxie/beego), via direct integration

## License

This project is licensed under the [Apache 2.0 license](https://github.com/casbin/casbin/blob/master/LICENSE).

## Contact

If you have any issues or feature requests, please contact us. PR is welcomed.
- https://github.com/casbin/casbin/issues
- hsluoyz@gmail.com
