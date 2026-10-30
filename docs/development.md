# In-Memory Renderer Development

## Prerequisites and commands

Use Go 1.26.8 and the repository Makefile:

```bash
make test
make test/race
make fmt
make lint
make lint/fix
make check
```

## Testing

Tests should verify deep-copy isolation on both input and output, source
selection, hook ordering, empty-object validation, and the default and
disabled content-hash behavior. Keep fixtures small and construct objects with
the Kubernetes unstructured helpers.

Concurrent rendering is covered by `TestRendererConcurrency`, which shares one
renderer across goroutines released through a start barrier and asserts that
every result is successful, carries the expected hook effects, and owns its
nested maps and slices independently of other results, the original sources,
and subsequent renders.

See [`design.md`](design.md) and [`../AGENTS.md`](../AGENTS.md).
