# AI Assistant Guide for k8s-manifest-kit/renderer-mem

## Quick Reference

This is the **Memory renderer** for the k8s-manifest-kit ecosystem. It provides programmatic rendering of pre-constructed Kubernetes objects that are already in memory. It's the simplest renderer with no I/O operations.

### Repository Structure
- `pkg/` - Main renderer implementation
- `docs/` - Architecture and development documentation

### Key Files
- `pkg/mem.go` - Main renderer (`New()`, `Process()`)
- `pkg/mem_option.go` - Functional options (`WithFilter()`, `WithTransformer()`, etc.)
- `pkg/mem_support.go` - Helper functions and validation
- `pkg/engine.go` - Convenience function (`NewEngine()`)

### Related Repositories
- `github.com/k8s-manifest-kit/engine` - Core engine and types
- `github.com/k8s-manifest-kit/pkg` - Shared utilities

## Common Tasks

### Understanding the Code

**Q: How does the renderer work?**
1. Sources contain pre-constructed `unstructured.Unstructured` objects
2. Objects are deep copied to prevent external mutations
3. Optional source annotations are added (only type, no path/file)
4. Results are filtered/transformed per pipeline configuration
5. Objects are returned (no caching needed)

**Q: What's the difference between `New()` and `NewEngine()`?**
- `New()` creates a `Renderer` implementing `types.Renderer`
- `NewEngine()` creates an `engine.Engine` with a single mem renderer (convenience)

**Q: Why use the mem renderer?**
Primary use cases:
- **Testing**: Pass pre-constructed objects without file I/O
- **Composition**: Combine objects from multiple sources
- **Programmatic generation**: Objects created dynamically in code
- **Mocking**: Simulate renderer behavior without external dependencies

### Making Changes

**Adding a renderer option:**
1. Add field to `RendererOptions` struct
2. Create `WithXxx()` function returning `RendererOption`
3. Add test coverage
4. Update documentation

**Adding validation:**
1. Update `Validate()` in `mem_support.go`
2. Add appropriate error types
3. Test error cases

### Testing

**Run tests:**
```bash
make test
```

**Test structure:**
- Unit tests in `pkg/*_test.go`
- Tests use actual `unstructured.Unstructured` objects
- Uses Gomega assertions (dot import)

**Key test files:**
- `mem_test.go` - Main renderer tests
- `engine_test.go` - NewEngine tests

### Debugging

**Common issues:**
1. **Empty objects**: Objects must have non-nil internal data
2. **Mutations**: Objects are deep copied, external changes won't affect rendered output
3. **Import paths**: Must use `github.com/k8s-manifest-kit/*`

**Useful debugging:**
```bash
# Run specific test
go test -v ./pkg -run TestRenderer

# Run with verbose output
go test -v ./...
```

## Architecture Notes

### Thread Safety
The renderer is thread-safe:
- Configuration is immutable after creation
- Deep copies prevent shared mutable state
- No external I/O or caching

### Deep Copying
Critical design element:
- All objects are deep copied before processing
- Prevents external code from modifying rendered objects
- Ensures isolation between renders

### No Caching
Unlike other renderers:
- No caching layer needed (objects already in memory)
- No expensive I/O to optimize
- Simplest renderer architecture

### Pipeline Integration
The renderer integrates with the three-level pipeline:
1. **Renderer-specific** (via `New()` options)
2. **Engine-level** (via `engine.New()` options)
3. **Render-time** (via `engine.Render()` options)

## Development Tips

1. **Follow established patterns** from other renderers
2. **Use functional options** for all configuration
3. **Document non-obvious behavior** in comments
4. **Test with realistic objects** in test code
5. **Check the linter** (`make lint`) - it's aggressive
6. **Keep imports organized** per `.golangci.yml` rules

## Code Review Checklist

When reviewing changes:
- [ ] Tests added for new functionality
- [ ] Error messages are clear and actionable
- [ ] Documentation updated (design.md, development.md)
- [ ] Follows Go conventions (parameter types, etc.)
- [ ] Thread safety considered
- [ ] Linter passes
- [ ] Imports use new k8s-manifest-kit paths

## Common Patterns

### Creating a renderer:
```go
r, err := mem.New(
    []mem.Source{{
        Objects: []unstructured.Unstructured{pod, service, deployment},
    }},
    mem.WithSourceAnnotations(true),
)
```

### Using NewEngine:
```go
e, err := mem.NewEngine(
    mem.Source{
        Objects: []unstructured.Unstructured{pod, service},
    },
)
```

### With filtering:
```go
r, err := mem.New(
    []mem.Source{{Objects: allObjects}},
    mem.WithFilter(gvk.Filter(corev1.SchemeGroupVersion.WithKind("Pod"))),
)
```

### Multiple sources:
```go
r, err := mem.New(
    []mem.Source{
        {Objects: coreObjects},
        {Objects: appObjects},
    },
)
```

## Key Differences from Other Renderers

**vs. YAML:**
- Mem: Objects in memory, no file I/O
- YAML: Loads from files, glob pattern matching

**vs. Kustomize:**
- Mem: Simple pass-through with deep copy
- Kustomize: Full kustomization processing with overlays

**vs. Helm:**
- Mem: No templating, static objects
- Helm: Full chart processing with values

**Use mem when:**
- Testing other parts of the system
- Objects are generated programmatically
- Composition from multiple sources
- No file I/O overhead desired

## Questions?

Check:
1. `docs/design.md` - Architecture and design decisions
2. `docs/development.md` - Development workflow
3. `pkg/*_test.go` - Usage examples
4. Parent repository documentation at github.com/k8s-manifest-kit

