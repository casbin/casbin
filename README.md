Casbin
====

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin)](https://goreportcard.com/report/github.com/casbin/casbin)
[![Build Status](https://travis-ci.org/casbin/casbin.svg?branch=master)](https://travis-ci.org/casbin/casbin)
[![Coverage Status](https://coveralls.io/repos/github/casbin/casbin/badge.svg?branch=master)](https://coveralls.io/github/casbin/casbin?branch=master)
[![Godoc](https://godoc.org/github.com/casbin/casbin?status.svg)](https://godoc.org/github.com/casbin/casbin)
[![Release](https://img.shields.io/github/release/casbin/casbin.svg)](https://github.com/casbin/casbin/releases/latest)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/casbin/lobby)
[![Patreon](https://img.shields.io/badge/patreon-donate-yellow.svg)](http://www.patreon.com/yangluo)
[![Sourcegraph Badge](https://sourcegraph.com/github.com/casbin/casbin/-/badge.svg)](https://sourcegraph.com/github.com/casbin/casbin?badge)

**News**: Casbin is also started to port to Java, contribution is welcomed. See: https://github.com/casbin/jcasbin

![casbin Logo](casbin-logo.png)

Casbin is a powerful and efficient open-source access control library for Golang projects. It provides support for enforcing authorization based on various [access control models](https://en.wikipedia.org/wiki/Computer_security_model).

## Table of contents

- [Supported models](#supported-models)
- [How it works?](#how-it-works)
- [Features](#features)
- [Installation](#installation)
- [Get started](#get-started)
- [Documentation](#documentation)
- [Policy persistence](#policy-persistence)
- [Examples](#examples)
- [Our adopters](#our-adopters)

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

## How it works?

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

## Documentation

See: [Our Wiki](https://github.com/casbin/casbin/wiki)

## Policy persistence

In Casbin, the policy storage is implemented as an adapter (aka middleware for Casbin). To keep light-weight, we don't put adapter code in the main library. A complete list of Casbin adapters is provided as below. Any 3rd-party contribution on a new adapter is welcomed, please inform us and I will put it in this list:)

Adapter | Type | Author | Description
----|------|----|----
[File Adapter (built-in)](https://github.com/casbin/casbin/wiki/Policy-persistence#file-adapter) | File | Casbin | Persistence for [.CSV (Comma-Separated Values)](https://en.wikipedia.org/wiki/Comma-separated_values) files
[Xorm Adapter](https://github.com/casbin/xorm-adapter) | ORM | Casbin | MySQL, PostgreSQL, TiDB, SQLite, SQL Server, Oracle are supported by [Xorm](https://github.com/go-xorm/xorm/)
[Gorm Adapter](https://github.com/casbin/gorm-adapter) | ORM | Casbin | MySQL, PostgreSQL, Sqlite3, SQL Server are supported by [Gorm](https://github.com/jinzhu/gorm/)
[MongoDB Adapter](https://github.com/casbin/mongodb-adapter) | NoSQL | Casbin | Persistence for [MongoDB](https://www.mongodb.com)
[Cassandra Adapter](https://github.com/casbin/cassandra-adapter) | NoSQL | Casbin | Persistence for [Apache Cassandra DB](http://cassandra.apache.org)
[Consul Adapter](https://github.com/ankitm123/consul-adapter) | KV store | [@ankitm123](https://github.com/ankitm123) | Persistence for [HashiCorp Consul](https://www.consul.io/)
[Redis Adapter](https://github.com/casbin/redis-adapter) | KV store | Casbin | Persistence for [Redis](https://redis.io/)
[Protobuf Adapter](https://github.com/casbin/protobuf-adapter) | Stream | Casbin | Persistence for [Google Protocol Buffers](https://developers.google.com/protocol-buffers/)
[RQLite Adapter](https://github.com/edomosystems/rqlite-adapter) | SQL | [EDOMO Systems](https://github.com/edomosystems) | Persistence for [RQLite](https://github.com/rqlite/rqlite/)
[PostgreSQL Adapter](https://github.com/going/casbin-postgres-adaptor) | SQL | [Going](https://github.com/going) | Persistence for [PostgreSQL](https://www.postgresql.org/)

For details of adapters, please refer to the documentation: https://github.com/casbin/casbin/wiki/Policy-persistence

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

## RoleManager

To use a custom RoleManager implementation.

```go

    type myCustomRoleManager struct {} // assumes the type satisfies the RoleManager interface

    func newRoleManager() rbac.RoleManagerConstructor {
        return func() rbac.RoleManager {
            return &myCustomRoleManager{}
        }
    }

    e := casbin.NewEnforcer("path/to/model.conf", "path/to/policy.csv")
    e.SetRoleManager(newRoleManager())
```

## How to use Casbin as a service?

- [Go-Simple-API-Gateway](https://github.com/Soontao/go-simple-api-gateway): A simple API gateway written by golang, supports for authentication and authorization
- [Casbin Server](https://github.com/casbin/casbin-server): Casbin as a Service via RESTful, only exposed permission checking API

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
- [Baa](https://github.com/go-baa/baa): An express Go web framework with routing, middleware, dependency injection and http context, via plugin: [authz](https://github.com/baa-middleware/authz)

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

## Donation

[![Patreon](https://hsluoyz.github.io/donation/patreon.png)](http://www.patreon.com/yangluo)

![Alipay](https://hsluoyz.github.io/donation/donate_alipay.png)
![Wechat](https://hsluoyz.github.io/donation/donate_weixin.png)
