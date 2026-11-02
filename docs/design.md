# In-Memory Renderer Design

## Source and ownership

`mem.Source` contains a slice of `unstructured.Unstructured` objects and
optional source-specific post-renderers.

`mem.New` retains the supplied `Source` values, including their object slices,
by reference. Construction is not a detached snapshot: callers own the source
objects and must not mutate them while the renderer is in use, including
concurrently with `Process` calls.

Every `Process` call deep-copies the selected objects before running its
pipeline, so each returned result owns its nested maps and slices
independently of other results, of the original sources, and of subsequent
renders. There is no output cache; each call renders again from the retained
inputs.

## Pipeline

Source selectors choose inputs before processing. Source post-renderers run on
each source result. After source results are combined, renderer-level filters,
transformers, and post-renderers run through the shared pipeline.

Render-time values are accepted for compatibility with the shared renderer
contract but do not change in-memory objects. Empty or invalid source objects
are validated according to the renderer's input checks.

Source annotations are disabled by default. Content hashes use the shared
`manifests.k8s-manifests-kit/content.hash` annotation and are enabled by
default.

### Option composition

Renderer options apply left to right. Functional options touch a single
field — `WithFilter`, `WithTransformer`, `WithPostRenderer`, and
`WithSourceSelector` append, and the boolean setters overwrite only their
own field — so they compose with anything around them. A
`RendererOptions` struct is stronger: it replaces the hook, selector, and
boolean fields outright (including with nil and false). This renderer has
no cache or function-map configuration, so a struct application preserves
nothing.

A hook before a struct survives only if the struct carries it; a hook after
a struct is kept:

```go
// The struct clears the filter added before it and disables hashing:
// its zero-valued Filters and ContentHash replace the previous state.
renderer, _ := mem.New(sources,
    mem.WithFilter(keepPods),
    mem.RendererOptions{SourceAnnotations: true},
    mem.WithContentHash(true), // re-enables what the struct turned off
)

// Hook after struct: the post-renderer is appended to the struct's state.
renderer, _ := mem.New(sources,
    mem.RendererOptions{SourceAnnotations: true, ContentHash: true},
    mem.WithPostRenderer(label),
)
```

A struct that wants hashing must set `ContentHash: true` explicitly; a
`WithContentHash(true)` after the struct achieves the same. These rules
govern the renderer's own options only — the engine's additive options
(`engine.WithFilter` and friends) follow the engine's rules, not these.

### Hash stages

Each `Process` call deep-copies the selected source objects and annotates
the copies before any hook runs: source annotations are applied first
(when enabled), then the content hash, then the live render origin — the
annotation phases always run in that order regardless of option order.
Source post-renderers and the renderer-level and engine hooks run last,
on already annotated objects.

Consequences callers can rely on: the hash covers the complete copied
object at hashing time, including any metadata the source objects already
carried — pre-existing annotations participate. It is not a hash of any
source byte representation, and it is not idempotent. Later source,
renderer, and engine processing does not automatically rehash, and hook
mutations need not match the stored digest. The full annotation phase
contract is the engine's
[annotation contract](../../engine/docs/design.md#annotations); this
renderer follows it at the points described above.

See [`../AGENTS.md`](../AGENTS.md) and [`development.md`](development.md).
