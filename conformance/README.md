# Casbin conformance cases

Shared test cases that every Casbin implementation should pass. They are generated from the Go implementation, which is the reference: each test records what Go's `EnforceEx()` returns for one request.

## Layout

- `data/`: model (`.conf`) and policy (`.csv`) files, copied from `examples/`.
- `cases/*.json`: one file per model and policy pair:

```json
{
  "model": "rbac_model.conf",
  "policy": "rbac_policy.csv",
  "tests": [
    {"request": ["alice", "data1", "read"], "allow": true, "explain": ["alice", "data1", "read"]},
    {"request": ["bob", "data1", "read"], "allow": false, "explain": []}
  ]
}
```

`model` and `policy` are file names in `data/`. Every request value is a string. `explain` is the policy rule returned by `EnforceEx()`, or `[]` when no rule decided the result.

## Running the cases

Load the enforcer from the two files with the default file adapter, call the implementation's `enforceEx` for every request, and compare both `allow` and `explain`. The Go test is [`conformance_test.go`](../conformance_test.go).

Other implementations read the cases from the directory in the `CASBIN_CONFORMANCE_DIR` environment variable and skip the test when it is not set. In CI, check out this directory from `apache/casbin`:

```yaml
- uses: actions/checkout@v4
  with:
    repository: apache/casbin
    path: casbin-conformance
    sparse-checkout: conformance
- run: <test command>
  env:
    CASBIN_CONFORMANCE_DIR: ${{ github.workspace }}/casbin-conformance/conformance
```

## Regenerating

Add a model and policy pair to `conformanceSources` in `conformance_test.go`, then run:

```shell
go test -run '^TestConformance$' -update-conformance .
```

Requests are generated from the values in the policy, and at most 300 are kept per case, half of them allowed where possible.
