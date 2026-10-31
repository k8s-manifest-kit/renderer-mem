package mem_test

import (
	"context"
	"sync"

	pkgtypes "github.com/k8s-manifest-kit/engine/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mem "github.com/k8s-manifest-kit/renderer-mem/pkg"

	. "github.com/onsi/gomega"
)

const (
	// concurrentWorkers is the number of goroutines rendering the same renderer
	// instance at once. It must be large enough to make overlapping calls likely
	// under the race detector.
	concurrentWorkers = 8

	// concurrencyRenderedAnnotation is set by a renderer-level post-renderer to
	// verify hook effects are visible in every concurrent result.
	concurrencyRenderedAnnotation = "mem-test.k8s-manifest-kit/rendered"

	// concurrencySourceHookAnnotation is set by a source-specific post-renderer
	// to verify per-source hooks run on the expected objects only.
	concurrencySourceHookAnnotation = "mem-test.k8s-manifest-kit/source-hook"
)

// renderConcurrently runs workers concurrent Process calls on the same renderer,
// releasing them together through a start barrier. Workers only collect their
// results; all assertions happen in the caller's goroutine.
func renderConcurrently(
	ctx context.Context,
	renderer *mem.Renderer,
	workers int,
) ([][]unstructured.Unstructured, []error) {
	results := make([][]unstructured.Unstructured, workers)
	errs := make([]error, workers)

	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			<-start

			objects, err := renderer.Process(ctx, nil)

			results[i] = objects
			errs[i] = err
		})
	}

	close(start)
	wg.Wait()

	return results, errs
}

// markConcurrentlyRendered is a concurrency-safe renderer-level post-renderer:
// it only mutates the per-render copies it receives.
func markConcurrentlyRendered(
	_ context.Context,
	objects []unstructured.Unstructured,
) ([]unstructured.Unstructured, error) {
	for i := range objects {
		annotations := objects[i].GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[concurrencyRenderedAnnotation] = "true"
		objects[i].SetAnnotations(annotations)
	}

	return objects, nil
}

// markPodSourceRendered is a concurrency-safe source-specific post-renderer
// applied only to the pod source.
func markPodSourceRendered(
	_ context.Context,
	objects []unstructured.Unstructured,
) ([]unstructured.Unstructured, error) {
	for i := range objects {
		annotations := objects[i].GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[concurrencySourceHookAnnotation] = "pod-source"
		objects[i].SetAnnotations(annotations)
	}

	return objects, nil
}

// concurrencyPod returns a pod with nested labels and container fields that
// later assertions can distinguish from mutations.
func concurrencyPod(g *WithT, name string) unstructured.Unstructured {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"app": "race-app"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:1.27"},
			},
		},
	}

	return toUnstructured(g, pod)
}

// concurrencyConfigMap returns a config map with a nested data map that later
// assertions can distinguish from mutations.
func concurrencyConfigMap(g *WithT, name string) unstructured.Unstructured {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"app": "race-app"},
		},
		Data: map[string]string{"port": "8080"},
	}

	return toUnstructured(g, configMap)
}

// toUnstructured converts a typed object using the Kubernetes helpers.
func toUnstructured(g *WithT, obj runtime.Object) unstructured.Unstructured {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	g.Expect(err).ToNot(HaveOccurred())

	return unstructured.Unstructured{Object: content}
}

// firstContainer returns the single container of a rendered pod so callers can
// assert or mutate its nested fields.
func firstContainer(g *WithT, pod *unstructured.Unstructured) map[string]any {
	containers, ok, err := unstructured.NestedSlice(pod.Object, "spec", "containers")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(containers).To(HaveLen(1))

	return containers[0].(map[string]any)
}

// expectConcurrentResult asserts the expected objects, nested source fields,
// and hook effects of a successful concurrent render.
func expectConcurrentResult(g *WithT, objects []unstructured.Unstructured) {
	g.Expect(objects).To(HaveLen(2))

	pod := findObject(objects, "Pod", "race-pod")
	g.Expect(pod).ToNot(BeNil())
	g.Expect(pod.GetLabels()).To(HaveKeyWithValue("app", "race-app"))
	g.Expect(firstContainer(g, pod)["image"]).To(Equal("nginx:1.27"))

	podAnnotations := pod.GetAnnotations()
	g.Expect(podAnnotations).To(HaveKeyWithValue(concurrencyRenderedAnnotation, "true"))
	g.Expect(podAnnotations).To(HaveKeyWithValue(concurrencySourceHookAnnotation, "pod-source"))
	g.Expect(podAnnotations).To(HaveKey(pkgtypes.AnnotationContentHash))

	configMap := findObject(objects, "ConfigMap", "race-config")
	g.Expect(configMap).ToNot(BeNil())
	g.Expect(configMap.GetLabels()).To(HaveKeyWithValue("app", "race-app"))

	port, ok, err := unstructured.NestedString(configMap.Object, "data", "port")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(port).To(Equal("8080"))

	configMapAnnotations := configMap.GetAnnotations()
	g.Expect(configMapAnnotations).To(HaveKeyWithValue(concurrencyRenderedAnnotation, "true"))
	g.Expect(configMapAnnotations).ToNot(HaveKey(concurrencySourceHookAnnotation))
	g.Expect(configMapAnnotations).To(HaveKey(pkgtypes.AnnotationContentHash))
}

// expectPodSourceUnchanged asserts the original pod source object was not
// affected by mutations applied to rendered results.
func expectPodSourceUnchanged(g *WithT, pod unstructured.Unstructured) {
	g.Expect(pod.GetLabels()).To(HaveKeyWithValue("app", "race-app"))
	g.Expect(firstContainer(g, &pod)["image"]).To(Equal("nginx:1.27"))
}

// expectConfigMapSourceUnchanged asserts the original config map source object
// was not affected by mutations applied to rendered results.
func expectConfigMapSourceUnchanged(g *WithT, configMap unstructured.Unstructured) {
	g.Expect(configMap.GetLabels()).To(HaveKeyWithValue("app", "race-app"))

	port, ok, err := unstructured.NestedString(configMap.Object, "data", "port")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(port).To(Equal("8080"))
}

// mutateNestedContent mutates nested maps and slices of a rendered result to
// prove that results own their data. It must only be called on results that no
// worker goroutine can still reach.
func mutateNestedContent(g *WithT, objects []unstructured.Unstructured) {
	pod := findObject(objects, "Pod", "race-pod")
	g.Expect(pod).ToNot(BeNil())

	labels := pod.GetLabels()
	labels["app"] = "mutated"
	pod.SetLabels(labels)

	containers, ok, err := unstructured.NestedSlice(pod.Object, "spec", "containers")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())

	containers[0].(map[string]any)["image"] = "mutated:latest"
	containers = append(containers, map[string]any{"name": "extra", "image": "extra:latest"})
	err = unstructured.SetNestedSlice(pod.Object, containers, "spec", "containers")
	g.Expect(err).ToNot(HaveOccurred())

	configMap := findObject(objects, "ConfigMap", "race-config")
	g.Expect(configMap).ToNot(BeNil())

	// GetLabels, NestedSlice, and NestedMap return copies, so every
	// mutation must be written back to actually touch the result.
	data, ok, err := unstructured.NestedMap(configMap.Object, "data")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())

	data["port"] = "mutated"
	err = unstructured.SetNestedMap(configMap.Object, data, "data")
	g.Expect(err).ToNot(HaveOccurred())
}

// findObject returns the object with the given kind and name, or nil.
func findObject(
	objects []unstructured.Unstructured,
	kind string,
	name string,
) *unstructured.Unstructured {
	for i := range objects {
		if objects[i].GetKind() == kind && objects[i].GetName() == name {
			return &objects[i]
		}
	}

	return nil
}
