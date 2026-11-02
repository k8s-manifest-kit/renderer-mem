package mem

import (
	"github.com/k8s-manifest-kit/engine/pkg/types"
	"github.com/k8s-manifest-kit/pkg/util/option"
)

// RendererOption is a generic option for RendererOptions.
type RendererOption = option.Option[RendererOptions]

// RendererOptions is a struct-based option that can set multiple renderer options at once.
//
// Options apply left to right: each option sees the state left by the ones
// before it. When RendererOptions is applied through ApplyTo, every field
// replaces its counterpart in the target — the hook and selector slices
// including when nil, and the booleans including when false. There is no
// cache or function-map configuration in this renderer, so nothing is
// preserved across a struct application.
//
// Functional options change only their own field or append to a single
// slice, so they compose with anything applied before or after. Note the
// difference from a fresh constructor: New starts from defaults with content
// hashing enabled, while applying a RendererOptions struct replaces the
// hook, selector, and boolean fields outright — a caller who wants content
// hashing after a struct must set ContentHash: true in the struct (or
// follow it with WithContentHash(true)).
type RendererOptions struct {
	// Filters are renderer-specific filters applied during Process().
	Filters []types.Filter

	// Transformers are renderer-specific transformers applied during Process().
	Transformers []types.Transformer

	// PostRenderers are renderer-specific post-renderers applied during Process().
	PostRenderers []types.PostRenderer

	// SourceSelectors are renderer-specific source selectors evaluated before rendering each source.
	SourceSelectors []SourceSelector

	// SourceAnnotations enables automatic addition of source tracking annotations.
	// Applying a RendererOptions struct replaces this field, including with false.
	SourceAnnotations bool

	// ContentHash enables automatic addition of a SHA-256 content hash annotation.
	// Default: true (enabled) in a fresh constructor; applying a
	// RendererOptions struct replaces this field, including with false, so a
	// struct that wants hashing must set it to true explicitly.
	ContentHash bool
}

// ApplyTo applies the renderer options to the target configuration.
//
// Application follows the composition rules documented on RendererOptions:
// every field replaces its target counterpart, including slices with nil
// and booleans with false.
func (opts RendererOptions) ApplyTo(target *RendererOptions) {
	target.Filters = opts.Filters
	target.Transformers = opts.Transformers
	target.PostRenderers = opts.PostRenderers
	target.SourceSelectors = opts.SourceSelectors
	target.SourceAnnotations = opts.SourceAnnotations
	target.ContentHash = opts.ContentHash
}

// WithFilter adds a renderer-specific filter to this Mem renderer's processing chain.
func WithFilter(f types.Filter) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.Filters = append(opts.Filters, f)
	})
}

// WithTransformer adds a renderer-specific transformer to this Mem renderer's processing chain.
func WithTransformer(t types.Transformer) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.Transformers = append(opts.Transformers, t)
	})
}

// WithPostRenderer adds a renderer-specific post-renderer to this Mem renderer's processing chain.
func WithPostRenderer(p types.PostRenderer) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.PostRenderers = append(opts.PostRenderers, p)
	})
}

// WithSourceSelector adds a source selector to this Mem renderer.
// Source selectors are evaluated before rendering each source. If any selector
// returns false, the source is skipped entirely.
func WithSourceSelector(s SourceSelector) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.SourceSelectors = append(opts.SourceSelectors, s)
	})
}

// WithSourceAnnotations enables or disables automatic addition of source tracking annotations.
func WithSourceAnnotations(enabled bool) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.SourceAnnotations = enabled
	})
}

// WithContentHash enables or disables automatic addition of a SHA-256 content hash annotation.
func WithContentHash(enabled bool) RendererOption {
	return option.FunctionalOption[RendererOptions](func(opts *RendererOptions) {
		opts.ContentHash = enabled
	})
}
