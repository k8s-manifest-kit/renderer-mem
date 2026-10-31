# In-Memory Renderer Development

## Prerequisites and commands

Use the Go version prescribed by [`go.mod`](../go.mod) and the repository
Makefile:

```bash
make test        # unit tests
make test/race   # unit tests with the race detector
make fmt         # formatting, separate from check
make lint
make lint/fix
make check       # lint and vulncheck
```

`make check` runs lint and vulnerability checks; it is not a substitute for
`make test` or `make test/race`. The default test suite is offline and there
is no live integration target.

## Code conventions

- Give each function parameter its type explicitly: `name string,
  namespace string`, not `name, namespace string`.
- Keep shared implementation helpers in `mem_support.go` and test-only support
  in the `_test.go` files; leave cohesive feature files such as `mem.go`
  intact rather than extracting every small helper.
- Wrap errors with operation context using `%w` when a cause exists, and check
  sentinels such as `ErrObjectEmpty` with `errors.Is`. Validation errors name
  the source index or object path needed for diagnosis.
- Tests live in the external `mem_test` package and exercise the public API
  with Gomega assertions, `t.Context()`, and Kubernetes unstructured helpers.

### Multiline strings

- Use one descriptive, standalone, package-level `const name =` declaration
  with a raw string literal for every multiline string: fixtures, manifests,
  templates, and expected output.
- Never embed multiline string literals in function or method bodies, and
  never declare them as function-local constants.
- Never group multiline constants in `const (...)`. Related single-line
  scalar constants can still be grouped, and single-line inline strings are
  unaffected.
- Keep large literal content separate from execution logic for readability.
  Preserve meaningful content indentation; the declaration layout must not
  introduce unintended bytes.
- A function may format the named template into a local variable; the
  prohibition concerns where the literal is declared, not where its string
  value is used.

```go
const fixtureYAML = `apiVersion: v1
kind: ConfigMap
`
```

## Ownership tests

`Process` deep-copies the selected source objects on every call, and tests
must prove that ownership claim with effective mutations:

- Kubernetes getters such as `GetLabels`, `NestedMap`, and `NestedSlice` can
  return detached copies. Write mutations back with `SetLabels`,
  `SetNestedMap`, and `SetNestedSlice` so they reach the actual result, as
  `mutateNestedContent` in `mem_concurrency_test.go` does.
- First assert that the mutation changed the result it was applied to, then
  verify that other results, the original source objects, and subsequent
  renders from the same renderer remain unaffected.
- To detect aliasing of a particular nested node, mutate that node — for
  example a container's `image` — rather than replacing its parent with a
  separately allocated map.

Construction is not a detached snapshot: `mem.New` retains the supplied
`Source` objects by reference, so tests must not mutate source objects after
handing them to the renderer. There is no output cache to reason about; every
call re-renders from the retained inputs.

## Concurrency

Concurrent rendering is covered by `TestRendererConcurrency`, which shares one
renderer across goroutines released together through a start barrier, joins
all workers before any assertion, and asserts on the test goroutine. Workers
own their results independently, and the post-renderer hooks used by the test
are concurrency-safe because they only mutate the per-call copies they
receive.

The renderer does not mutate its own configuration during `Process`, but any
caller-supplied hook — post-renderers, filters, transformers, or source
selectors — must be concurrency-safe when the same renderer is used by
overlapping calls; the library provides no automatic synchronization of user
code.

See [`design.md`](design.md) and [`../AGENTS.md`](../AGENTS.md).
