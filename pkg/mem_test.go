package mem_test

import (
	"testing"

	"github.com/k8s-manifest-kit/engine/pkg/filter/meta/gvk"
	"github.com/k8s-manifest-kit/engine/pkg/transformer/meta/labels"
	pkgtypes "github.com/k8s-manifest-kit/engine/pkg/types"
	jqmatcher "github.com/lburgazzoli/gomega-matchers/pkg/matchers/jq"
	"github.com/onsi/gomega/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

func TestRenderer(t *testing.T) {

	// Test objects
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"app":       "test-app",
				"component": "frontend",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-config",
			Labels: map[string]string{
				"app":       "test-app",
				"component": "frontend",
			},
		},
		Data: map[string]string{
			"config.yaml": "port: 8080",
		},
	}

	tests := []struct {
		name          string
		objects       []runtime.Object
		opts          []mem.RendererOption
		expectedCount int
		validation    types.GomegaMatcher
	}{
		{
			name:          "should return empty list for no objects",
			objects:       []runtime.Object{},
			expectedCount: 0,
			validation:    nil,
		},
		{
			name:          "should return single object unchanged",
			objects:       []runtime.Object{pod},
			expectedCount: 1,
			validation: And(
				jqmatcher.Match(`.kind == "Pod"`),
				jqmatcher.Match(`.metadata.name == "test-pod"`),
				jqmatcher.Match(`.metadata.labels["app"] == "test-app"`),
				jqmatcher.Match(`.metadata.labels["component"] == "frontend"`),
			),
		},
		{
			name:          "should return multiple objects unchanged",
			objects:       []runtime.Object{pod, configMap},
			expectedCount: 2,
			validation: Or(
				And(
					jqmatcher.Match(`.kind == "Pod"`),
					jqmatcher.Match(`.metadata.name == "test-pod"`),
				),
				And(
					jqmatcher.Match(`.kind == "ConfigMap"`),
					jqmatcher.Match(`.metadata.name == "test-config"`),
				),
			),
		},
		{
			name:    "should apply filters",
			objects: []runtime.Object{pod, configMap},
			opts: []mem.RendererOption{
				mem.WithFilter(gvk.Filter(corev1.SchemeGroupVersion.WithKind("Pod"))),
			},
			expectedCount: 1,
			validation: And(
				jqmatcher.Match(`.kind == "Pod"`),
				jqmatcher.Match(`.metadata.name == "test-pod"`),
			),
		},
		{
			name:    "should apply transformers",
			objects: []runtime.Object{pod},
			opts: []mem.RendererOption{
				mem.WithTransformer(labels.Set(map[string]string{
					"managed-by": "mem-renderer",
					"env":        "test",
				})),
			},
			expectedCount: 1,
			validation: And(
				jqmatcher.Match(`.kind == "Pod"`),
				jqmatcher.Match(`.metadata.labels["managed-by"] == "mem-renderer"`),
				jqmatcher.Match(`.metadata.labels["env"] == "test"`),
				jqmatcher.Match(`.metadata.labels["app"] == "test-app"`),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			// Convert typed objects to unstructured inline
			unstructuredObjects := make([]unstructured.Unstructured, len(tt.objects))
			for i, obj := range tt.objects {
				unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				g.Expect(err).ToNot(HaveOccurred())

				unstructuredObjects[i] = unstructured.Unstructured{Object: unstr}
			}

			renderer, err := mem.New([]mem.Source{{Objects: unstructuredObjects}}, tt.opts...)
			g.Expect(err).ToNot(HaveOccurred())

			objects, err := renderer.Process(t.Context(), nil)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(objects).To(HaveLen(tt.expectedCount))

			if tt.validation != nil {
				for _, obj := range objects {
					g.Expect(obj.Object).To(tt.validation)
				}
			}
		})
	}
}

func TestMetricsIntegration(t *testing.T) {

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "metrics-pod",
		},
	}

	// Metrics are now observed at the engine level, not in the renderer
	// This test verifies that renderers work without metrics in context
	t.Run("should work without metrics context", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)

		renderer, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{
				{Object: unstrPod},
			},
		}})
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).To(HaveLen(1))
	})

	t.Run("should implement Name() method", func(t *testing.T) {
		g := NewWithT(t)
		renderer, err := mem.New([]mem.Source{{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(renderer.Name()).To(Equal("mem"))
	})
}

func TestSourceAnnotations(t *testing.T) {

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}

	t.Run("should add source annotations when enabled", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New(
			[]mem.Source{{
				Objects: []unstructured.Unstructured{
					{Object: unstrPod},
				},
			}},
			mem.WithSourceAnnotations(true),
		)
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).Should(HaveLen(1))

		// Verify source annotations are present
		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).Should(HaveKeyWithValue(pkgtypes.AnnotationSourceType, "mem"))
		// Mem renderer should not have path or file annotations
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationSourcePath))
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationSourceFile))
	})

	t.Run("should not add source annotations when disabled", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{
				{Object: unstrPod},
			},
		}})
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).Should(HaveLen(1))

		// Verify no source annotations are present
		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationSourceType))
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationSourcePath))
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationSourceFile))
	})
}
