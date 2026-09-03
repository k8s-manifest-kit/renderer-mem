# In-Memory Renderer Development

## Prerequisites and commands

Use Go 1.26.8 and the repository Makefile:

```bash
make test
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

See [`design.md`](design.md) and [`../AGENTS.md`](../AGENTS.md).
