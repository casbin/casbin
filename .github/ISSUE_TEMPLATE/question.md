---
name: Question
about: Questions like "Why this model and policy don't work as expected?"
title: "[Question]"
labels: question
assignees: hsluoyz

---

**Want to prioritize this issue? Try:**

[![issuehunt-to-marktext](https://github.com/BoostIO/issuehunt-materials/raw/master/v1/issuehunt-button-v1.svg)](https://issuehunt.io/r/casbin/casbin)

------

**What's your scenario? What do you want to achieve?**
Your answer here

**Your model:**

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

**Your policy:**

```
p, alice, data1, read
p, bob, data2, write
p, data2_admin, data2, read
p, data2_admin, data2, write

g, alice, data2_admin
```

**Your request(s):**

```
alice, data2, read ---> false (expected: true)
```
