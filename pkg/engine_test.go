package mem_test

import (
	"errors"
	"testing"

	engine "github.com/k8s-manifest-kit/engine/pkg"
	"github.com/k8s-manifest-kit/engine/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

func TestNewEngine(t *testing.T) {

	t.Run("should create engine with mem renderer", func(t *testing.T) {
		g := NewWithT(t)

		pod := &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
		}

		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		e, err := mem.NewEngine(mem.Source{
			Objects: []unstructured.Unstructured{
				{Object: unstrPod},
			},
		})

		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(e).ShouldNot(BeNil())

		// Verify it can render
		objects, err := e.Render(t.Context())
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(objects).To(HaveLen(1))
	})

	t.Run("should return error for invalid source", func(t *testing.T) {
		g := NewWithT(t)

		// Empty object (invalid)
		e, err := mem.NewEngine(mem.Source{
			Objects: []unstructured.Unstructured{
				{Object: nil},
			},
		})

		g.Expect(err).Should(HaveOccurred())
		g.Expect(e).Should(BeNil())
	})

	t.Run("should work with empty objects list", func(t *testing.T) {
		g := NewWithT(t)

		e, err := mem.NewEngine(mem.Source{
			Objects: []unstructured.Unstructured{},
		})

		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(e).ShouldNot(BeNil())

		// Verify it returns empty list
		objects, err := e.Render(t.Context())
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(objects).To(BeEmpty())
	})
}

// TestEngineTypedNilRenderer verifies that engine construction accepts the
// typed-nil built-in — Name is nil-safe — and that rendering returns the
// nil-receiver sentinel without panic.
func TestEngineTypedNilRenderer(t *testing.T) {
	g := NewWithT(t)

	var renderer *mem.Renderer

	e, err := engine.New(engine.WithRenderer(renderer))
	g.Expect(err).ToNot(HaveOccurred(), "engine construction accepts the typed-nil built-in")

	objects, err := e.Render(t.Context())

	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, types.ErrRendererNil)).To(BeTrue())
	g.Expect(objects).To(BeNil())
}
