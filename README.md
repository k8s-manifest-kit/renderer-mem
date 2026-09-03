# In-Memory Renderer

`renderer-mem` renders a supplied slice of Kubernetes objects. It is useful
for programmatically constructed manifests, tests, and composing renderers in
an engine pipeline.

## Installation

```bash
go get github.com/k8s-manifest-kit/renderer-mem
```

## Quick start

```go
e, err := mem.NewEngine(mem.Source{
    Objects: []unstructured.Unstructured{object},
})
if err != nil {
    return err
}

objects, err := e.Render(ctx)
```

Input objects and returned objects are deep-copied. Source selectors,
post-renderers, source annotations, and content hashes are supported. Values
are accepted by the shared render API but do not alter in-memory objects.

See [`docs/design.md`](docs/design.md), [`docs/development.md`](docs/development.md),
and [`AGENTS.md`](AGENTS.md).

## License

Apache License 2.0. See [LICENSE](LICENSE).
