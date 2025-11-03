# Memory Renderer Development Guide

## Setup

### Prerequisites

- Go 1.24 or later
- Make
- golangci-lint

### Getting Started

```bash
# Clone and navigate
cd /path/to/k8s-manifest-kit/renderer-mem

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint
```

## Project Structure

```
renderer-mem/
├── pkg/
│   ├── mem.go              # Main renderer implementation
│   ├── mem_option.go       # Functional options
│   ├── mem_support.go      # Helper functions
│   ├── mem_test.go         # Tests
│   ├── engine.go           # NewEngine convenience
│   └── engine_test.go      # NewEngine tests
├── docs/
│   ├── design.md          # Architecture documentation
│   └── development.md     # This file
├── .golangci.yml          # Linter configuration
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── README.md              # Project overview
```

## Coding Conventions

### Go Style

Follow standard Go conventions plus:
- Each function parameter has its own type declaration
- Use multiline formatting for functions with 3+ parameters
- Prefer explicit types when they aid readability

### Error Handling

- Return errors as the last parameter
- Use `fmt.Errorf` with `%w` verb for wrapping
- Handle errors at appropriate abstraction level
- Provide clear, actionable error messages

Example:
```go
func (h *sourceHolder) Validate() error {
    for i := range h.Objects {
        if len(h.Objects[i].Object) == 0 {
            return fmt.Errorf("%w at index %d", ErrObjectEmpty, i)
        }
    }
    return nil
}
```

### Testing with Gomega

Use vanilla Gomega assertions:

```go
import . "github.com/onsi/gomega"

func TestExample(t *testing.T) {
    g := NewWithT(t)
    result, err := someFunction()
    
    g.Expect(err).ShouldNot(HaveOccurred())
    g.Expect(result).Should(HaveLen(3))
}
```

### Documentation

- Comments explain *why*, not *what*
- Focus on non-obvious behavior and edge cases
- Skip boilerplate docstrings unless they add value
- Document public APIs thoroughly

## Development Workflow

### Making Changes

1. **Write Tests First**: Add test cases for new functionality
2. **Implement**: Make minimal changes to fulfill requirements
3. **Run Tests**: `make test`
4. **Run Linter**: `make lint`
5. **Fix Issues**: Address any linter warnings

### Adding New Features

#### Adding a Renderer Option

1. Add field to `RendererOptions` struct in `mem_option.go`
2. Create `WithXxx()` function
3. Add test coverage
4. Update documentation

Example:
```go
// In mem_option.go
type RendererOptions struct {
    // ... existing fields ...
    NewFeature bool
}

func WithNewFeature(enabled bool) RendererOption {
    return util.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
        opts.NewFeature = enabled
    })
}
```

#### Adding Validation

1. Update `Validate()` in `mem_support.go`
2. Add appropriate error types if needed
3. Handle in renderer logic
4. Add test coverage for validation failures

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific test
go test -v ./pkg -run TestRenderer

# Run specific sub-test
go test -v ./pkg -run "TestRenderer/should_return_single"

# Run benchmarks
make bench
```

### Test Structure

Tests use `unstructured.Unstructured` objects directly:

```go
func TestExample(t *testing.T) {
    g := NewWithT(t)
    
    pod := &corev1.Pod{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "v1",
            Kind: "Pod",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name: "test-pod",
        },
    }
    
    unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
    g.Expect(err).ToNot(HaveOccurred())
    
    renderer, err := mem.New([]mem.Source{{
        Objects: []unstructured.Unstructured{{Object: unstrPod}},
    }})
    g.Expect(err).ToNot(HaveOccurred())
    
    objects, err := renderer.Process(t.Context(), nil)
    g.Expect(err).ToNot(HaveOccurred())
    g.Expect(objects).To(HaveLen(1))
}
```

### Test Coverage

Key test files:
- `mem_test.go`: Main renderer tests
  - Object pass-through
  - Deep copying
  - Filters and transformers
  - Error cases
  - Source annotations
  - Benchmarks
- `engine_test.go`: NewEngine convenience function tests

### Writing Good Tests

1. **Test behavior, not implementation**
2. **Use descriptive test names**: `"should return deep copies of objects"`
3. **One assertion focus per test**
4. **Use table-driven tests for similar cases**
5. **Test error paths, not just happy paths**

## Benchmarking

Benchmark key operations:

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkSpecific -benchmem ./pkg
```

Focus areas for benchmarking:
- Deep copy overhead
- Filter/transformer application
- Scaling with object count
- Memory allocations

## Linting

The project uses an aggressive linter configuration:

```bash
# Run linter
make lint

# Auto-fix issues
make lint/fix

# Format code
make fmt
```

### Common Linter Issues

1. **Import ordering**: Use `gci` formatter (runs automatically with `make fmt`)
2. **Error wrapping**: Always use `%w` verb
3. **Variable naming**: Avoid single-letter names outside loops
4. **Dot imports**: Only for Gomega test assertions

## Debugging

### Common Issues

1. **Empty objects**
   - Check that Object field is non-nil
   - Verify conversion from typed to unstructured succeeded

2. **Unexpected mutations**
   - Remember: objects are deep copied
   - External changes won't affect rendered output

3. **Import path errors**
   - Use `github.com/k8s-manifest-kit/*` paths
   - Run `go mod tidy` after changing imports

### Useful Debugging Commands

```bash
# Verbose test output
go test -v ./pkg -run TestName

# Print test with race detector
go test -race ./...

# Check module dependencies
go mod graph | grep k8s-manifest-kit

# Verify imports
go list -m all
```

## Dependencies

### Core Dependencies
- `k8s.io/apimachinery` - Kubernetes types
- `github.com/k8s-manifest-kit/engine` - Core engine
- `github.com/k8s-manifest-kit/pkg` - Shared utilities

### Test Dependencies
- `github.com/onsi/gomega` - Assertions
- `github.com/lburgazzoli/gomega-matchers` - JQ matchers
- `k8s.io/api` - Kubernetes API types

### Updating Dependencies

```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get github.com/k8s-manifest-kit/engine@latest
go mod tidy
```

## Code Review Guidelines

### Before Submitting

- [ ] Tests pass locally
- [ ] Linter passes
- [ ] Documentation updated
- [ ] Error messages are clear
- [ ] Follows coding conventions

### Review Checklist

- [ ] Code follows established patterns
- [ ] Tests cover new functionality
- [ ] Error handling is appropriate
- [ ] Documentation is clear
- [ ] No unnecessary complexity
- [ ] Thread safety considered
- [ ] Deep copying maintained

## Common Patterns

### Creating Test Objects

```go
pod := &corev1.Pod{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "v1",
        Kind: "Pod",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name: "test-pod",
        Labels: map[string]string{"app": "test"},
    },
}

unstrPod, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
```

### Testing Error Cases

```go
t.Run("should return error for empty object", func(t *testing.T) {
    g := NewWithT(t)
    _, err := mem.New([]mem.Source{
        {Objects: []unstructured.Unstructured{{Object: nil}}},
    })
    
    g.Expect(err).Should(HaveOccurred())
    g.Expect(err.Error()).To(ContainSubstring("object is empty"))
})
```

### Testing Deep Copying

```go
t.Run("should return deep copies", func(t *testing.T) {
    g := NewWithT(t)
    original := unstructured.Unstructured{Object: map[string]interface{}{"key": "value"}}
    
    renderer, _ := mem.New([]mem.Source{{Objects: []unstructured.Unstructured{original}}})
    result, _ := renderer.Process(t.Context(), nil)
    
    // Modify original
    original.Object["key"] = "modified"
    
    // Result should be unchanged
    g.Expect(result[0].Object["key"]).To(Equal("value"))
})
```

## Resources

- [Kubernetes API reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Unstructured documentation](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured)
- [Gomega documentation](https://onsi.github.io/gomega/)

## Questions?

Check:
1. [CLAUDE.md](../CLAUDE.md) - Quick reference for common tasks
2. [Design documentation](design.md) - Architecture details
3. Test files - Usage examples
4. Parent organization at github.com/k8s-manifest-kit

