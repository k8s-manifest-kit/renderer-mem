// Package mem provides a memory-based renderer for Kubernetes manifests.
// It handles rendering of unstructured objects that are already in memory.
package mem

import (
	"context"
	"fmt"

	"github.com/k8s-manifest-kit/engine/pkg/pipeline"
	"github.com/k8s-manifest-kit/engine/pkg/types"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const rendererType = "mem"

// Source represents the input for a memory-based rendering operation.
type Source struct {
	// Objects contains pre-constructed Kubernetes manifests to pass through.
	// Useful for testing, composition, or when objects are already in memory.
	Objects []unstructured.Unstructured

	// PostRenderers are source-specific post-renderers applied to this source's output
	// before combining with other sources.
	PostRenderers []types.PostRenderer
}

// Renderer handles memory-based rendering operations.
// It implements types.Renderer for objects that are already in memory.
type Renderer struct {
	inputs []*sourceHolder
	opts   RendererOptions
}

// New creates a new memory-based renderer with the given inputs and options.
func New(inputs []Source, opts ...RendererOption) (*Renderer, error) {
	rendererOpts := RendererOptions{
		Filters:      make([]types.Filter, 0),
		Transformers: make([]types.Transformer, 0),
		ContentHash:  true,
	}

	for _, opt := range opts {
		opt.ApplyTo(&rendererOpts)
	}

	// Wrap sources in holders and validate
	holders := make([]*sourceHolder, len(inputs))
	for i := range inputs {
		holders[i] = &sourceHolder{
			Source: inputs[i],
		}
		if err := holders[i].Validate(); err != nil {
			return nil, fmt.Errorf("invalid source at index %d: %w", i, err)
		}
	}

	r := &Renderer{
		inputs: holders,
		opts:   rendererOpts,
	}

	return r, nil
}

// Process implements types.Renderer by returning the objects that were provided during construction.
// Render-time values are ignored by the memory renderer as objects are already constructed.
func (r *Renderer) Process(ctx context.Context, _ types.Values) ([]unstructured.Unstructured, error) {
	allObjects := make([]unstructured.Unstructured, 0)

	for _, holder := range r.inputs {
		selected, err := pipeline.ApplySourceSelectors(ctx, holder.Source, r.opts.SourceSelectors)
		if err != nil {
			return nil, fmt.Errorf("source selector error in mem renderer: %w", err)
		}

		if !selected {
			continue
		}

		sourceObjects := make([]unstructured.Unstructured, 0, len(holder.Objects))

		for _, obj := range holder.Objects {
			objCopy := obj.DeepCopy()

			if r.opts.SourceAnnotations {
				annotations := objCopy.GetAnnotations()
				if annotations == nil {
					annotations = make(map[string]string)
				}

				annotations[types.AnnotationSourceType] = rendererType

				objCopy.SetAnnotations(annotations)
			}

			sourceObjects = append(sourceObjects, *objCopy)
		}

		if r.opts.ContentHash {
			for i := range sourceObjects {
				types.SetContentHash(&sourceObjects[i])
			}
		}

		sourceObjects, err = pipeline.ApplyPostRenderers(ctx, sourceObjects, holder.PostRenderers)
		if err != nil {
			return nil, fmt.Errorf("source post-renderer error in mem renderer: %w", err)
		}

		allObjects = append(allObjects, sourceObjects...)
	}

	chain := types.BuildPostRendererChain(r.opts.Filters, r.opts.Transformers, r.opts.PostRenderers)

	result, err := pipeline.ApplyPostRenderers(ctx, allObjects, chain)
	if err != nil {
		return nil, fmt.Errorf("renderer post-renderer error in mem renderer: %w", err)
	}

	return result, nil
}

// Name returns the renderer type identifier.
func (r *Renderer) Name() string {
	return rendererType
}
