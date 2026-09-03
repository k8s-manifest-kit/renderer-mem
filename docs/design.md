# In-Memory Renderer Design

## Source and ownership

`mem.Source` contains a slice of `unstructured.Unstructured` objects and
optional source-specific post-renderers. The renderer deep-copies objects when
it accepts a source and when it returns results, so callers and the renderer do
not share mutable object state.

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

See [`../AGENTS.md`](../AGENTS.md) and [`development.md`](development.md).
