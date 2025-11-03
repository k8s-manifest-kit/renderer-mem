# renderer-mem

Memory-based renderer for the k8s-manifest-kit ecosystem.

Part of the [k8s-manifest-kit](https://github.com/k8s-manifest-kit) organization.

## Overview

The Memory renderer is the simplest renderer in the k8s-manifest-kit ecosystem. It provides programmatic rendering of pre-constructed Kubernetes objects that are already in memory, with no file I/O, no templates, and no overlays. Perfect for testing, composition, and programmatic object generation.

## Installation

```bash
go get github.com/k8s-manifest-kit/renderer-mem
```

## Quick Start

```go
package main

import (
    "context"
    
    mem "github.com/k8s-manifest-kit/renderer-mem/pkg"
    
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime"
)

func main() {
    // Create typed Kubernetes objects
    pod := &corev1.Pod{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "v1",
            Kind: "Pod",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name: "my-pod",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "app", Image: "nginx:latest"},
            },
        },
    }
    
    // Convert to unstructured
    unstrPod, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
    
    // Create engine with mem renderer
    e, err := mem.NewEngine(mem.Source{
        Objects: []unstructured.Unstructured{
            {Object: unstrPod},
        },
    })
    if err != nil {
        panic(err)
    }
    
    // Render objects
    objects, err := e.Render(context.Background())
    if err != nil {
        panic(err)
    }
    
    // Process objects...
}
```

## Features

- **Zero I/O**: No file reading or network calls - objects are already in memory
- **Deep Copying**: All objects are deep copied to prevent external mutations
- **Filtering & Transformation**: Apply filters and transformers at render time
- **Source Tracking**: Optional annotations to track object origins
- **Thread-Safe**: Safe for concurrent use
- **Testing-Friendly**: Perfect for unit tests and mocking

## Use Cases

### Testing
Pass pre-constructed objects without file fixtures:
```go
renderer, _ := mem.New([]mem.Source{{Objects: testObjects}})
results, _ := renderer.Process(ctx, nil)
```

### Composition
Combine objects from multiple sources:
```go
renderer, _ := mem.New([]mem.Source{
    {Objects: coreObjects},
    {Objects: appObjects},
    {Objects: monitoringObjects},
})
```

### Programmatic Generation
Work with dynamically created objects:
```go
objects := generateObjects() // Your custom logic
e, _ := mem.NewEngine(mem.Source{Objects: objects})
```

### Mocking
Simulate renderer behavior in tests:
```go
mockRenderer, _ := mem.New([]mem.Source{{Objects: mockData}})
```

## Documentation

- [Design Documentation](docs/design.md) - Architecture and design decisions
- [Development Guide](docs/development.md) - Development workflow and conventions
- [CLAUDE.md](CLAUDE.md) - AI assistant reference guide

## Key Differences from Other Renderers

| Feature | Memory | YAML | Kustomize | Helm |
|---------|--------|------|-----------|------|
| Input | In-memory objects | Files | Kustomization | Charts |
| I/O | None | File reading | File reading | Network + files |
| Templates | No | No | No | Yes |
| Caching | Not needed | Yes | Yes | Yes |
| Complexity | Minimal | Low | Medium | High |

## Contributing

Contributions are welcome! Please see our [contributing guidelines](https://github.com/k8s-manifest-kit/docs/blob/main/CONTRIBUTING.md).

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

