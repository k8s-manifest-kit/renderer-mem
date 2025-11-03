package mem

import (
	"fmt"

	engine "github.com/k8s-manifest-kit/engine/pkg"
)

// NewEngine creates an Engine configured with a single memory renderer.
// This is a convenience function for simple in-memory rendering scenarios.
//
// Example:
//
//	e, _ := mem.NewEngine(
//	    mem.Source{
//	        Objects: []unstructured.Unstructured{...},
//	    },
//	)
//	objects, _ := e.Render(ctx)
func NewEngine(source Source, opts ...RendererOption) (*engine.Engine, error) {
	renderer, err := New([]Source{source}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mem renderer: %w", err)
	}

	e, err := engine.New(engine.WithRenderer(renderer))
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	return e, nil
}
