package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/iter8ctl/utils"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func getK8sClientWithTargetFromFile(filePath string) (client.Client, error) {
	data, err := ioutil.ReadFile(utils.CompletePath("testdata", filePath))
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
	scheme := runtime.NewScheme()
	etc3.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(u).Build(), nil
}

type myk8s struct {
	c client.Client
}

func (k *myk8s) GetClient() (client.Client, error) {
	return k.c, nil
}

// used in the following two tests
var expectedVersionInfo = &etc3.VersionInfo{
	Baseline: etc3.VersionDetail{
		Name: "default",
		Tags: &map[string]string{"revision": "my-model-predictor-default-wl2cv"},
	},
	Candidates: []etc3.VersionDetail{
		{
			Name: "canary",
			Tags: &map[string]string{"revision": "my-model-predictor-default-zwjbq"},
			WeightObjRef: &v1.ObjectReference{
				Kind:       "InferenceService",
				Namespace:  "default",
				Name:       "my-model",
				APIVersion: "serving.kubeflow.org/v1beta1",
				FieldPath:  "/spec/predictor/canaryTrafficPercent",
			},
		},
	},
}

// Mocking os.Exit function
type testOS struct{}

func (t *testOS) Exit(code int) {
	if code > 0 {
		panic(fmt.Sprintf("Exiting with error code %v", code))
	} else {
		panic("Normal exit")
	}
}

// initTestOS registers the mock OS struct (testOS) defined above
func initTestOS() {
	osExiter = &testOS{}
}

func TestMain(t *testing.T) {
	c, err := getK8sClientWithTargetFromFile("canaryv1beta1.json")
	if err != nil {
		t.Fatal("Cannot get k8s client with target from file")
	}
	exp := etc3.NewExperiment("myexp", "default").
		WithTarget("default/my-model").
		WithStrategy(etc3.StrategyTypeCanary).
		Build()
	err = c.Create(context.Background(), exp)
	if err != nil {
		t.Fatal("Cannot populate fake cluster with experiment", err)
	}
	k8s = &myk8s{c}
	os.Args = []string{"./handler", "start"}
	os.Setenv("EXPERIMENT_NAME", "myexp")
	os.Setenv("EXPERIMENT_NAMESPACE", "default")
	main()
	os.Unsetenv("EXPERIMENT_NAME")
	os.Unsetenv("EXPERIMENT_NAMESPACE")
	c.Get(context.Background(), client.ObjectKeyFromObject(exp), exp)
	assert.Equal(t, expectedVersionInfo, exp.Spec.VersionInfo)
}

func TestMainNoArgs(t *testing.T) {
	initTestOS()
	k8s = &myk8s{fake.NewClientBuilder().Build()}
	os.Args = []string{"./handler"}
	assert.Panics(t, func() { main() })
}

func TestMainInvalidArgs(t *testing.T) {
	initTestOS()
	k8s = &myk8s{fake.NewClientBuilder().Build()}
	os.Args = []string{"./handler", "invalid"}
	assert.Panics(t, func() { main() })
}

func TestMainCannotGetExperiment(t *testing.T) {
	initTestOS()
	k8s = &myk8s{fake.NewClientBuilder().Build()}
	os.Args = []string{"./handler", "start"}
	assert.Panics(t, func() { main() })
}

func TestMainSingleVersionFinish(t *testing.T) {
	c, err := getK8sClientWithTargetFromFile("canaryv1beta1.json")
	if err != nil {
		t.Fatal("Cannot get k8s client with target from file")
	}
	exp := etc3.NewExperiment("myexp", "default").
		WithTarget("default/my-model").
		WithStrategy(etc3.StrategyTypePerformance).
		Build()
	err = c.Create(context.Background(), exp)
	if err != nil {
		t.Fatal("Cannot populate fake cluster with experiment", err)
	}
	k8s = &myk8s{c}
	os.Args = []string{"./handler", "finish"}
	os.Setenv("EXPERIMENT_NAME", "myexp")
	os.Setenv("EXPERIMENT_NAMESPACE", "default")
	assert.PanicsWithValue(t, "Normal exit", func() { main() })
	os.Unsetenv("EXPERIMENT_NAME")
	os.Unsetenv("EXPERIMENT_NAMESPACE")
}

func TestMainFinishError(t *testing.T) {
	c, err := getK8sClientWithTargetFromFile("canaryv1beta1.json")
	if err != nil {
		t.Fatal("Cannot get k8s client with target from file")
	}
	exp := etc3.NewExperiment("myexp", "default").
		WithTarget("default/my-model").
		WithStrategy(etc3.StrategyTypeCanary).
		Build()
	err = c.Create(context.Background(), exp)
	if err != nil {
		t.Fatal("Cannot populate fake cluster with experiment", err)
	}
	k8s = &myk8s{c}
	os.Args = []string{"./handler", "finish"}
	os.Setenv("EXPERIMENT_NAME", "myexp")
	os.Setenv("EXPERIMENT_NAMESPACE", "default")
	assert.Panics(t, func() { main() })
	os.Unsetenv("EXPERIMENT_NAME")
	os.Unsetenv("EXPERIMENT_NAMESPACE")
}
