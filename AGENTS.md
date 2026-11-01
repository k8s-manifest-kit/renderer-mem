# Agent Guide: renderer-mem

`renderer-mem` passes through pre-constructed Kubernetes objects held in memory. It performs no file or network I/O and is useful for tests, composition, and programmatic generation.

## Documentation

- [README](README.md) — overview and usage.
- [Design](docs/design.md) — behavior and tradeoffs.
- [Development](docs/development.md) — workflow and tests.

## Public API

The package is imported from `github.com/k8s-manifest-kit/renderer-mem/pkg`.

- `mem.New([]mem.Source{...}, opts...)` creates a renderer.
- `mem.NewEngine(source, opts...)` creates an `engine.Engine` for one source.
- `Source` contains `Objects` and optional source-specific `PostRenderers`.
- Options include filters, transformers, post-renderers, source selectors, source annotations, and content hashes.

`mem.New` retains the supplied source objects by reference; construction is
not a detached snapshot, so do not mutate source objects while the renderer is
in use. Each `Process` call deep-copies the selected objects before
processing, so every result owns its data independently. Render-time values
are ignored because objects are already constructed. Source annotations, when
enabled, add only the renderer type; content hashes are enabled by default.
Empty objects are rejected during construction.

## Development

Run commands from this directory:

```bash
make test        # unit tests
make test/race   # unit tests with the race detector
make fmt         # formatting, separate from check
make lint
make lint/fix
make check       # lint and vulncheck
```

Use the Go version prescribed by `go.mod`. Use unstructured objects in tests,
Gomega assertions, `t.Context()`, and explicit tests for deep-copy isolation,
selectors, post-renderers, annotations, hashes, and validation errors.
Keep tests focused and simple: use the smallest object fixture that proves
the behavior, and add helpers or extra test files only when a concrete
behavior or regression requires them. Ownership and concurrency test rules
are described in [Development](docs/development.md).

