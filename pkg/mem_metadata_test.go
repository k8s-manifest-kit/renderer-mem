package mem_test

import (
	"testing"

	pkgtypes "github.com/k8s-manifest-kit/engine/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

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
		g.Expect(annotations).Should(HaveKeyWithValue(
			pkgtypes.AnnotationRenderOrigin,
			pkgtypes.RenderOriginLive,
		))
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

func TestContentHash(t *testing.T) {

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
				{Name: "nginx", Image: "nginx:latest"},
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
		},
		Data: map[string]string{"key": "value"},
	}

	t.Run("should add content hash annotation by default", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{{Object: unstrPod}},
		}})
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).Should(HaveLen(1))

		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).Should(HaveKey(pkgtypes.AnnotationContentHash))
		g.Expect(annotations[pkgtypes.AnnotationContentHash]).Should(MatchRegexp("^sha256:[0-9a-f]{64}$"))
	})

	t.Run("should not add content hash when disabled", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New(
			[]mem.Source{{
				Objects: []unstructured.Unstructured{{Object: unstrPod}},
			}},
			mem.WithContentHash(false),
		)
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).Should(HaveLen(1))

		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).ShouldNot(HaveKey(pkgtypes.AnnotationContentHash))
	})

	t.Run("different objects should have different hashes", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())
		unstrCM, err := runtime.DefaultUnstructuredConverter.ToUnstructured(configMap)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{
				{Object: unstrPod},
				{Object: unstrCM},
			},
		}})
		g.Expect(err).ToNot(HaveOccurred())

		objects, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(objects).Should(HaveLen(2))

		hash0 := objects[0].GetAnnotations()[pkgtypes.AnnotationContentHash]
		hash1 := objects[1].GetAnnotations()[pkgtypes.AnnotationContentHash]
		g.Expect(hash0).ShouldNot(Equal(hash1))
	})

	t.Run("hash should be stable across renders", func(t *testing.T) {
		g := NewWithT(t)
		unstrPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())

		renderer, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{{Object: unstrPod}},
		}})
		g.Expect(err).ToNot(HaveOccurred())

		result1, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())

		result2, err := renderer.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())

		hash1 := result1[0].GetAnnotations()[pkgtypes.AnnotationContentHash]
		hash2 := result2[0].GetAnnotations()[pkgtypes.AnnotationContentHash]
		g.Expect(hash1).ShouldNot(BeEmpty())
		g.Expect(hash1).Should(Equal(hash2))
	})

	t.Run("hash should change when content changes", func(t *testing.T) {
		g := NewWithT(t)

		podV1, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		g.Expect(err).ToNot(HaveOccurred())
		r1, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{{Object: podV1}},
		}})
		g.Expect(err).ToNot(HaveOccurred())
		objects1, err := r1.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())

		podModified := pod.DeepCopy()
		podModified.Spec.Containers[0].Image = "nginx:1.27"
		podV2, err := runtime.DefaultUnstructuredConverter.ToUnstructured(podModified)
		g.Expect(err).ToNot(HaveOccurred())
		r2, err := mem.New([]mem.Source{{
			Objects: []unstructured.Unstructured{{Object: podV2}},
		}})
		g.Expect(err).ToNot(HaveOccurred())
		objects2, err := r2.Process(t.Context(), nil)
		g.Expect(err).ToNot(HaveOccurred())

		hash1 := objects1[0].GetAnnotations()[pkgtypes.AnnotationContentHash]
		hash2 := objects2[0].GetAnnotations()[pkgtypes.AnnotationContentHash]
		g.Expect(hash1).ShouldNot(BeEmpty())
		g.Expect(hash2).ShouldNot(BeEmpty())
		g.Expect(hash1).ShouldNot(Equal(hash2))
	})
}
