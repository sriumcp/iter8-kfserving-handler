package v1beta1

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/iter8ctl/utils"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
		Version: "v1beta1",
	})
	return fake.NewClientBuilder().WithObjects(u).Build()
}

func getK8sClientWithTargetFromFile(filePath string) (client.Client, error) {
	data, err := ioutil.ReadFile(utils.CompletePath("../testdata", filePath))
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}
	err = json.Unmarshal(data, &u.Object)
	if err != nil {
		return nil, err
	} // we have gotten our unstructured object so far.
	return fake.NewClientBuilder().WithObjects(u).Build(), nil
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
	namespace, name, err := getNN("myns/myname")
	assert.Equal(t, "myns", namespace)
	assert.Equal(t, "myname", name)
	assert.NoError(t, err)

	namespace, name, err = getNN("v1beta1/myns/myname")
	assert.Equal(t, "myns", namespace)
	assert.Equal(t, "myname", name)
	assert.NoError(t, err)

	namespace, name, err = getNN("v1alpha2/myns/myname")
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
	targ.SetK8sClient(c).Fetch("v1beta1/myns/myname")
	assert.NoError(t, targ.err)
}

func TestFetchNonExisting(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	targ := TargetBuilder()
	targ.SetK8sClient(c).Fetch("myns/myname")
	assert.Error(t, targ.err)
}

func TestInitializeTrafficSplit(t *testing.T) {
	c, err := getK8sClientWithTargetFromFile("canaryv1beta1.json")
	if err != nil {
		t.Fatal("Cannot get k8s client with target from file")
	}
	targ := TargetBuilder()
	exp := etc3.NewExperiment("myexp", "default").
		WithTarget("default/my-model").
		WithStrategy(etc3.StrategyTypeCanary).
		Build()
	targ.exp = experiment.Builder(exp)
	t.Log("targetRef: ", targ.exp.GetTargetRef())
	targ.SetK8sClient(c).Fetch("default/my-model").InitializeTrafficSplit()
	assert.NoError(t, targ.err)

	i, b, err := unstructured.NestedInt64(targ.infService.Object, "spec", "predictor", "canaryTrafficPercent")
	assert.True(t, b)
	assert.Equal(t, int64(1), i)
	assert.NoError(t, err)
}

func TestGetConditions(t *testing.T) {
	c, err := getK8sClientWithTargetFromFile("canaryv1beta1.json")
	if err != nil {
		t.Fatal("Cannot get k8s client with target from file")
	}
	targ := TargetBuilder()
	exp := etc3.NewExperiment("myexp", "default").
		WithTarget("default/my-model").
		WithStrategy(etc3.StrategyTypeCanary).
		Build()
	targ.exp = experiment.Builder(exp)
	t.Log("targetRef: ", targ.exp.GetTargetRef())
	targ.SetK8sClient(c).Fetch("default/my-model").InitializeTrafficSplit()
	assert.NoError(t, targ.err)
	cond, err := GetConditions(targ)
	assert.NoError(t, err)
	assert.NotEmpty(t, cond)
	ready, _ := target.GetCondition(cond, "Ready")
	assert.Equal(t, "True", ready)
}
