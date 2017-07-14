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

Casbin is a powerful and efficient open-source access control library for Golang projects. It provides support for enforcing authorization based on various [access control models](https://en.wikipedia.org/wiki/Computer_security_model).

## Supported models

1. [**ACL (Access Control List)**](https://en.wikipedia.org/wiki/Access_control_list)
2. **ACL with [superuser](https://en.wikipedia.org/wiki/Superuser)**
3. **ACL without users**: especially useful for systems that don't have authentication or user log-ins.
3. **ACL without resources**: some scenarios may target for a type of resources instead of an individual resource by using permissions like ``write-article``, ``read-log``. It doesn't control the access to a specific article or log.
4. **[RBAC (Role-Based Access Control)](https://en.wikipedia.org/wiki/Role-based_access_control)**
5. **RBAC with resource roles**: both users and resources can have roles (or groups) at the same time.
6. **RBAC with domains/tenants**: users can have different role sets for different domains/tenants.
7. **[ABAC (Attribute-Based Access Control)](https://en.wikipedia.org/wiki/Attribute-Based_Access_Control)**: syntax sugar like ``resource.Owner`` can be used to get the attribute for a resource.
8. **[RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer)**: supports paths like ``/res/*``, ``/res/:id`` and HTTP methods like ``GET``, ``POST``, ``PUT``, ``DELETE``.
9. **Deny-override**: both allow and deny authorizations are supported, deny overrides the allow.
10. **Priority**: the policy rules can be prioritized like firewall rules.

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

1. New a Casbin enforcer with a model file and a policy file:

```go
e := casbin.NewEnforcer("path/to/model.conf", "path/to/policy.csv")
```

Note: you can also initialize an enforcer with policy in DB instead of file, see [Persistence](#persistence) section for details.

2. Add an enforcement hook into your code right before the access happens:

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

3. Besides the static policy file, Casbin also provides API for permission management at run-time. For example, You can get all the roles assigned to a user as below:

```go
roles := e.GetRoles("alice")
```

Note: we provide two sets of APIs to manage permissions:

- [Management API](https://github.com/casbin/casbin/blob/master/management_api.go): the primitive API that provides full support for Casbin policy management.
- [RBAC API](https://github.com/casbin/casbin/blob/master/rbac_api.go): a more friendly API for RBAC. This API is a subset of Management API. The RBAC users could use this API to simplify the code.

4. Please refer to the ``_test.go`` files for more usage.

## Syntax for models

See: [Model.md](https://github.com/casbin/casbin/blob/master/Model.md)

## Persistence

In Casbin, the policy storage is implemented as an adapter (aka middleware for Casbin). To keep light-weight, we don't put adapter code in the main library. A complete list of Casbin adapters is provided as below. Any 3rd-party contribution on a new adapter is welcomed, please inform us and I will put it in this list:)

Adapter | Type | Author | Description
----|------|----|----
[File Adapter (built-in)](#file-adapter) | File | Casbin | Persistence for [.CSV (Comma-Separated Values)](https://en.wikipedia.org/wiki/Comma-separated_values) files
[MySQL Adapter](https://github.com/casbin/mysql-adapter) | RDBMS | Casbin | Persistence for [MySQL](https://www.mysql.com)
[Cassandra Adapter](https://github.com/casbin/cassandra-adapter) | NoSQL | Casbin | Persistence for [Apache Cassandra DB](http://cassandra.apache.org)
[Consul Adapter](https://github.com/ankitm123/consul-adapter) | KV store | [@ankitm123](https://github.com/ankitm123) | Persistence for [HashiCorp Consul](https://www.consul.io/)
[Redis Adapter](https://github.com/ankitm123/redis-adapter) | KV store | [@ankitm123](https://github.com/ankitm123) | Persistence for [Redis](https://redis.io/)
[Protobuf Adapter](https://github.com/casbin/protobuf-adapter) | Stream | Casbin | Persistence for [Google Protocol Buffers](https://developers.google.com/protocol-buffers/)
[Xorm Adapter](https://github.com/casbin/xorm-adapter) | ORM | Casbin | MySQL, PostgreSQL, TiDB, SQLite, SQL Server, Oracle are supported by [Xorm](https://github.com/go-xorm/xorm/)

All adapters should implement the [Adapter interface](https://github.com/casbin/casbin/blob/master/persist/adapter.go) by providing two methods:``LoadPolicy(model model.Model) error`` and ``SavePolicy(model model.Model) error``. And as a convention, the adapter should be able to automatically create a database named ``casbin``  if it doesn't exist and use it for policy storage.

Note: Unlike the policy, the model can be loaded from a CONF file only. Because we think the model is not a frequently modified part at run-time, so we don't implement an API to save the model into a file or database.

### File adapter

Below shows how to initialize an enforcer from the built-in file adapter:

```go
e := casbin.NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
```

This is the same with:

```go
a := persist.NewFileAdapter("examples/basic_policy.csv")
e := casbin.NewEnforcer("examples/basic_model.conf", a)
```

### MySQL adapter

Below shows how to initialize an enforcer from MySQL database. it connects to a MySQL DB on 127.0.0.1:3306 with root and blank password.

```go
a := mysqladapter.NewDBAdapter("mysql", "root:@tcp(127.0.0.1:3306)/")
e := casbin.NewEnforcer("examples/basic_model.conf", a)
```

### Use your own storage adapter

You can use your own adapter like below:

```go
a := yourpackage.YourNewAdapter(your_params)
e := casbin.NewEnforcer("examples/basic_model.conf", a)
```

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

## What if I don't want to reply on files (.conf, .csv)?

The policy rules can be loaded and saved with various adapters, so it doesn't rely on ``.CSV`` files necessarily. The model can be initialized from code instead of using ``.CONF`` file. Here's an example for the RBAC model:

```go
// Initialize the model from Go code.
m := casbin.NewModel()
m.AddDef("r", "r", "sub, obj, act")
m.AddDef("p", "p", "sub, obj, act")
m.AddDef("g", "g", "_, _")
m.AddDef("e", "e", "some(where (p.eft == allow))")
m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

// Load the policy rules from the .CSV file adapter.
// Replace it with your adapter to avoid files.
a := persist.NewFileAdapter("examples/rbac_policy.csv")

// Create the enforcer.
e := casbin.NewEnforcer(m, a)
```

Or you can just get the model from one single string:

```go
// Initialize the model from a string.
text :=
`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
m := NewModel(text)

// Load the policy rules from the .CSV file adapter.
// Replace it with your adapter to avoid files.
a := persist.NewFileAdapter("examples/rbac_policy.csv")

// Create the enforcer.
e := casbin.NewEnforcer(m, a)
```

These two ways are equivalent with the common usage:

[examples/rbac_model.conf](https://github.com/casbin/casbin/blob/master/examples/rbac_model.conf):

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

```go
e := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
```

## Error handling

Error or panic may happen when you use Casbin for reasons like:

1. Invalid syntax in model file (.conf).
2. Invalid syntax in policy file (.csv).
3. Custom error from storage adapters, e.g., MySQL fails to connect.
4. Casbin's bug.

There are five main functions you may need to care about for error or panic:

Function | Behavior on error | What if I want error instead of panic?
----|------|----
[NewEnforcer()](https://godoc.org/github.com/casbin/casbin#NewEnforcer) | Cause panic | Please use [NewEnforcerSafe()](https://godoc.org/github.com/casbin/casbin#NewEnforcerSafe)
[LoadModel()](https://godoc.org/github.com/casbin/casbin#Enforcer.LoadModel) | Cause panic | Please use [LoadModelSafe()](https://godoc.org/github.com/casbin/casbin#Enforcer.LoadModelSafe)
[LoadPolicy()](https://godoc.org/github.com/casbin/casbin#Enforcer.LoadPolicy) | Return error | N/A
[SavePolicy()](https://godoc.org/github.com/casbin/casbin#Enforcer.SavePolicy) | Return error | N/A
[Enforce()](https://godoc.org/github.com/casbin/casbin#Enforcer.Enforce) | Cause panic | Please use [EnforceSafe()](https://godoc.org/github.com/casbin/casbin#Enforcer.EnforceSafe)

Note: ``NewEnforcer()`` calls ``LoadModel()`` and ``LoadPolicy()`` inside. So you don't have to call the latter two calls when using ``NewEnforcer()``.

### Why not just return error for all the functions?

The author believes that Golang error is very unfriendly for developers to debug, because it's only a string and has no faulty call stack information. The panic can show call stack and integrate well with Golang IDEs. Most of the Casbin users are developers too. And they usually don't write the Casbin model or policy correctly at the first time (which causes error or panic). So they can benefit from the advantages of panic. However, there are still many people who favor error more than panic. So we provide the ``xxxSafe()`` functions which just wrap the ``xxx()`` functions by translating panic into error. You can just use ``xxx()`` or ``xxxSafe()`` functions based on personal tastes.

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
Priority | [priority_model.conf](https://github.com/casbin/casbin/blob/master/examples/priority_model.conf)  | [priority_policy.csv](https://github.com/casbin/casbin/blob/master/examples/priority_policy.csv)

## Our adopters

### Web frameworks

- [Beego](https://github.com/astaxie/beego): An open-source, high-performance web framework for Go, via built-in plugin: [plugins/authz](https://github.com/astaxie/beego/blob/master/plugins/authz)
- [Caddy](https://github.com/mholt/caddy): Fast, cross-platform HTTP/2 web server with automatic HTTPS, via plugin: [caddy-authz](https://github.com/casbin/caddy-authz)
- [Gin](https://github.com/gin-gonic/gin): A HTTP web framework featuring a Martini-like API with much better performance, via plugin: [authz](https://github.com/gin-contrib/authz)
- [Revel](https://github.com/revel/revel): A high productivity, full-stack web framework for the Go language, via plugin: [revel-authz](https://github.com/casbin/revel-authz)
- [Echo](https://github.com/labstack/echo): High performance, minimalist Go web framework, via plugin: [echo-authz](https://github.com/labstack/echo-contrib/tree/master/casbin) (thanks to [@xqbumu](https://github.com/xqbumu))
- [Iris](https://github.com/kataras/iris): The fastest web framework for Go in (THIS) Earth. HTTP/2 Ready-To-GO, via plugin: [casbin](https://github.com/iris-contrib/middleware/tree/master/casbin) (thanks to [@hiveminded](https://github.com/hiveminded))
- [Negroni](https://github.com/urfave/negroni): Idiomatic HTTP Middleware for Golang, via plugin: [negroni-authz](https://github.com/casbin/negroni-authz)
- [Tango](https://github.com/lunny/tango): Micro & pluggable web framework for Go, via plugin: [authz](https://github.com/tango-contrib/authz)
- [Chi](https://github.com/pressly/chi): A lightweight, idiomatic and composable router for building HTTP services, via plugin: [chi-authz](https://github.com/casbin/chi-authz)
- [Macaron](https://github.com/go-macaron/macaron): A high productive and modular web framework in Go, via plugin: [authz](https://github.com/go-macaron/authz)
- [DotWeb](https://github.com/devfeel/dotweb): Simple and easy go web micro framework, via plugin: [authz](https://github.com/devfeel/middleware/tree/master/authz)

### Others

- [Docker](https://github.com/docker/docker): The world's leading software container platform, via plugin: [casbin-authz-plugin](https://github.com/casbin/casbin-authz-plugin) ([recommended by Docker](https://docs.docker.com/engine/extend/legacy_plugins/#authorization-plugins))
- [pybbs-go](https://github.com/tomoya92/pybbs-go): A simple BBS with fine-grained permission management based on [Beego](https://github.com/astaxie/beego), via direct integration
- [Zenpress](https://github.com/insionng/zenpress): A CMS system written in Golang, via direct integration

## License

This project is licensed under the [Apache 2.0 license](https://github.com/casbin/casbin/blob/master/LICENSE).

## Contact

If you have any issues or feature requests, please contact us. PR is welcomed.
- https://github.com/casbin/casbin/issues
- hsluoyz@gmail.com
- Tencent QQ group: [546057381](//shang.qq.com/wpa/qunwpa?idkey=8ac8b91fc97ace3d383d0035f7aa06f7d670fd8e8d4837347354a31c18fac885)
