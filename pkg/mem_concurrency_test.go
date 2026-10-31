package mem_test

import (
	"testing"

	pkgtypes "github.com/k8s-manifest-kit/engine/pkg/types"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

// TestRendererConcurrency verifies that concurrent calls to the same renderer
// return successful, independent results: each render deep-copies the source
// objects, so nested maps and slices are never shared between results, with
// the original sources, or with subsequent renders.
func TestRendererConcurrency(t *testing.T) {
	g := NewWithT(t)

	pod := concurrencyPod(g, "race-pod")
	configMap := concurrencyConfigMap(g, "race-config")

	sources := []mem.Source{
		{
			Objects:       []unstructured.Unstructured{pod},
			PostRenderers: []pkgtypes.PostRenderer{markPodSourceRendered},
		},
		{
			Objects: []unstructured.Unstructured{configMap},
		},
	}

	renderer, err := mem.New(sources, mem.WithPostRenderer(markConcurrentlyRendered))
	g.Expect(err).ToNot(HaveOccurred())

	results, errs := renderConcurrently(t.Context(), renderer, concurrentWorkers)

	for i := range concurrentWorkers {
		g.Expect(errs[i]).ToNot(HaveOccurred())
		expectConcurrentResult(g, results[i])
	}

	// Mutating one result must not affect any other concurrent result...
	mutateNestedContent(g, results[0])

	for i := 1; i < concurrentWorkers; i++ {
		expectConcurrentResult(g, results[i])
	}

	// ...nor the original source objects...
	expectPodSourceUnchanged(g, pod)
	expectConfigMapSourceUnchanged(g, configMap)

	// ...nor a subsequent render from the same renderer.
	fresh, err := renderer.Process(t.Context(), nil)
	g.Expect(err).ToNot(HaveOccurred())
	expectConcurrentResult(g, fresh)
}
