package v1alpha2

import (
	"testing"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getK8sClientWithMyTarget() client.Client {
	// Using a unstructured object.
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "myname",
			"namespace": "myns",
		},
		"spec": map[string]interface{}{},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "serving.kubeflow.org",
		Kind:    "InferenceService",
		Version: "v1alpha2",
	})
	return fake.NewClientBuilder().WithObjects(u).Build()
}

func TestTargetBuilder(t *testing.T) {
	x := TargetBuilder()
	assert.NoError(t, x.err)
}

func TestSetK8sClient(t *testing.T) {
	x := TargetBuilder()
	scheme := runtime.NewScheme()
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	x.SetK8sClient(c)
	assert.Equal(t, x.k8sclient, c)
}

func TestSetExperiment(t *testing.T) {
	x := TargetBuilder()
	e := &experiment.Experiment{}
	x.SetExperiment(e)
	assert.Equal(t, x.exp, e)
}

func TestGetNN(t *testing.T) {
	namespace, name, err := getNN("v1alpha2/myns/myname")
	assert.Equal(t, "myns", namespace)
	assert.Equal(t, "myname", name)
	assert.NoError(t, err)

	namespace, name, err = getNN("v1alpha1/myns/myname")
	assert.Error(t, err)

	namespace, name, err = getNN("myname")
	assert.Error(t, err)

	namespace, name, err = getNN("a/b/c/myname")
	assert.Error(t, err)

	namespace, name, err = getNN("")
	assert.Error(t, err)
}

func TestFetch(t *testing.T) {
	c := getK8sClientWithMyTarget()
	targ := TargetBuilder()
	targ.SetK8sClient(c).Fetch("v1alpha2/myns/myname")
	assert.NoError(t, targ.err)
}

func TestFetchNonExisting(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	targ := TargetBuilder()
	targ.SetK8sClient(c).Fetch("v1alpha2/myns/myname")
	assert.Error(t, targ.err)
}
