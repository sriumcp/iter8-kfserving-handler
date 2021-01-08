package experiment

import (
	"os"
	"testing"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildMyExp() *etc3.Experiment {
	return etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypeCanary).
		WithRequestCount("request-count").
		Build()
}

func getK8sClientWithMyExp() client.Client {
	crScheme := k8sruntime.NewScheme()
	err := etc3.AddToScheme(crScheme)
	if err != nil {
		panic("Error while adding to etc3's v1alpha1 to new scheme")
	}
	exp := buildMyExp()
	return fake.NewClientBuilder().WithScheme(crScheme).WithObjects(exp).Build()
}

func TestBuilder(t *testing.T) {
	exp := buildMyExp()
	e := Builder(exp)
	assert.NotEmpty(t, e)
}

func TestGetExperiment(t *testing.T) {
	c := getK8sClientWithMyExp()
	err := os.Setenv("EXPERIMENT_NAME", "myexp")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("EXPERIMENT_NAMESPACE", "myns")
	if err != nil {
		t.Fatal(err)
	}
	_, err = GetExperiment(c)
	assert.NoError(t, err)
}

func TestGetExperimentNonExistant(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	_, err := GetExperiment(c)
	assert.Error(t, err)
}

func TestExpGetTargetRef(t *testing.T) {
	exp := etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypeCanary).
		WithRequestCount("request-count").
		Build()
	e := &Experiment{exp}
	assert.Equal(t, "target", e.GetTargetRef())
}

func TestIsSingleVersion(t *testing.T) {
	exp := etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypeCanary).
		Build()
	e := Builder(exp)
	assert.False(t, e.IsSingleVersion())

	exp = etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypePerformance).
		Build()
	e = Builder(exp)
	assert.True(t, e.IsSingleVersion())

	exp = etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypeBlueGreen).
		Build()
	e = Builder(exp)
	assert.False(t, e.IsSingleVersion())

	exp = etc3.NewExperiment("myexp", "myns").
		WithTarget("target").
		WithStrategy(etc3.StrategyTypeAB).
		Build()
	e = Builder(exp)
	assert.False(t, e.IsSingleVersion())
}
