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

See [`../AGENTS.md`](../AGENTS.md) and [`development.md`](development.md).
