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

Each input object is deep-copied before processing. Render-time values are ignored because objects are already constructed. Source annotations, when enabled, add only the renderer type; content hashes are enabled by default. Empty objects are rejected during construction.

## Development

Run commands from this directory:

```bash
make test
make fmt
make lint
make lint/fix
make check
```

Use unstructured objects in tests, Gomega assertions, `t.Context()`, and explicit tests for deep-copy isolation, selectors, post-renderers, annotations, hashes, and validation errors.

