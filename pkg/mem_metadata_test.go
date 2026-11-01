package mem_test

import (
	"testing"

	pkgtypes "github.com/k8s-manifest-kit/engine/pkg/types"
	"github.com/k8s-manifest-kit/pkg/util/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

// annotatedMemObject returns a ConfigMap carrying its own annotations,
// including stale kit-managed keys, to characterize how renderer annotation
// phases interact with pre-existing metadata.
func annotatedMemObject(name string) unstructured.Unstructured {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"existing":                      "annotation",
				pkgtypes.AnnotationContentHash:  "sha256:stale",
				pkgtypes.AnnotationRenderOrigin: "stale",
				pkgtypes.AnnotationSourcePath:   "seeded/path.yaml",
				pkgtypes.AnnotationSourceFile:   "seeded/file.yaml",
			},
		},
		Data: map[string]string{"key": "value"},
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(configMap)
	if err != nil {
		panic(err)
	}

	return unstructured.Unstructured{Object: unstructuredMap}
}

// cleanMemObject returns a ConfigMap without any annotations.
func cleanMemObject(name string) unstructured.Unstructured {
	configMap := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Data:       map[string]string{"key": "value"},
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(configMap)
	if err != nil {
		panic(err)
	}

	return unstructured.Unstructured{Object: unstructuredMap}
}

// renderMemObjects processes the given objects with the renderer options.
func renderMemObjects(
	t *testing.T,
	objects []unstructured.Unstructured,
	opts ...mem.RendererOption,
) []unstructured.Unstructured {
	t.Helper()

	g := NewWithT(t)

	renderer, err := mem.New([]mem.Source{{Objects: objects}}, opts...)
	g.Expect(err).ToNot(HaveOccurred())

	result, err := renderer.Process(t.Context(), nil)
	g.Expect(err).ToNot(HaveOccurred())

	return result
}

// TestAnnotationApplication characterizes how the memory renderer applies
// type-only source metadata, content hashes, and live origins.
func TestAnnotationApplication(t *testing.T) {

	t.Run("exact annotation map with hash over pre-existing metadata before live origin", func(t *testing.T) {
		g := NewWithT(t)

		input := annotatedMemObject("annotated")
		objects := renderMemObjects(t, []unstructured.Unstructured{input}, mem.WithSourceAnnotations(true))

		g.Expect(objects).To(HaveLen(1))

		expected := input.DeepCopy()
		expected.SetAnnotations(map[string]string{
			"existing":                      "annotation",
			pkgtypes.AnnotationContentHash:  "sha256:stale",
			pkgtypes.AnnotationRenderOrigin: "stale",
			pkgtypes.AnnotationSourcePath:   "seeded/path.yaml",
			pkgtypes.AnnotationSourceFile:   "seeded/file.yaml",
			pkgtypes.AnnotationSourceType:   "mem",
		})

		g.Expect(objects[0].GetAnnotations()).To(Equal(map[string]string{
			"existing":                      "annotation",
			pkgtypes.AnnotationSourceType:   "mem",
			pkgtypes.AnnotationSourcePath:   "seeded/path.yaml",
			pkgtypes.AnnotationSourceFile:   "seeded/file.yaml",
			pkgtypes.AnnotationContentHash:  k8s.ContentHash(expected),
			pkgtypes.AnnotationRenderOrigin: pkgtypes.RenderOriginLive,
		}))
	})

	t.Run("source annotations without content hash leave no hash key and no path or file", func(t *testing.T) {
		g := NewWithT(t)

		objects := renderMemObjects(t,
			[]unstructured.Unstructured{cleanMemObject("no-hash")},
			mem.WithSourceAnnotations(true),
			mem.WithContentHash(false),
		)

		g.Expect(objects).To(HaveLen(1))

		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).To(HaveLen(2))
		g.Expect(annotations).To(HaveKeyWithValue(pkgtypes.AnnotationSourceType, "mem"))
		g.Expect(annotations).To(HaveKeyWithValue(pkgtypes.AnnotationRenderOrigin, pkgtypes.RenderOriginLive))
		g.Expect(annotations).NotTo(HaveKey(pkgtypes.AnnotationContentHash))
		g.Expect(annotations).NotTo(HaveKey(pkgtypes.AnnotationSourcePath))
		g.Expect(annotations).NotTo(HaveKey(pkgtypes.AnnotationSourceFile))
	})

	t.Run("disabled annotations hash over pre-existing metadata and inject no type or origin", func(t *testing.T) {
		g := NewWithT(t)

		input := annotatedMemObject("bare")
		objects := renderMemObjects(t, []unstructured.Unstructured{input})

		g.Expect(objects).To(HaveLen(1))

		annotations := objects[0].GetAnnotations()
		g.Expect(annotations).To(HaveKeyWithValue(
			pkgtypes.AnnotationContentHash,
			k8s.ContentHash(input.DeepCopy()),
		))
		g.Expect(annotations).To(HaveKeyWithValue("existing", "annotation"))
		g.Expect(annotations).To(HaveKeyWithValue(pkgtypes.AnnotationSourcePath, "seeded/path.yaml"))
		g.Expect(annotations).To(HaveKeyWithValue(pkgtypes.AnnotationSourceFile, "seeded/file.yaml"))
		// The stale origin is pre-existing metadata, not an injected one.
		g.Expect(annotations).To(HaveKeyWithValue(pkgtypes.AnnotationRenderOrigin, "stale"))
		g.Expect(annotations).NotTo(HaveKey(pkgtypes.AnnotationSourceType))
	})

	t.Run("caller objects remain untouched and results are deeply independent", func(t *testing.T) {
		g := NewWithT(t)

		input := annotatedMemObject("isolated")
		original := input.DeepCopy()

		first := renderMemObjects(t, []unstructured.Unstructured{input}, mem.WithSourceAnnotations(true))

		// Mutate nested content and the annotation map of the result.
		first[0].GetAnnotations()["mutated"] = "result-state"
		g.Expect(unstructured.SetNestedField(first[0].Object, "mutated", "data", "key")).To(Succeed())

		second := renderMemObjects(t, []unstructured.Unstructured{input}, mem.WithSourceAnnotations(true))

		g.Expect(input).To(Equal(*original))
		g.Expect(second[0].GetAnnotations()).NotTo(HaveKey("mutated"))
		g.Expect(second[0].GetAnnotations()).To(HaveKeyWithValue(pkgtypes.AnnotationSourceType, "mem"))
		g.Expect(second[0].GetAnnotations()).To(HaveKeyWithValue(pkgtypes.AnnotationContentHash,
			first[0].GetAnnotations()[pkgtypes.AnnotationContentHash]))
		data, found, err := unstructured.NestedString(second[0].Object, "data", "key")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(found).To(BeTrue())
		g.Expect(data).To(Equal("value"))
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
