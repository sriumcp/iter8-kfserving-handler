package v1alpha2

import (
	"context"
	"errors"
	"strings"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Target is an enhancement of KFServing v1alpha2 InferenceService.
type Target struct {
	err        error
	infService *unstructured.Unstructured
	exp        *experiment.Experiment
	k8sclient  client.Client
}

// TargetBuilder returns an initial v1beta1 target struct pointer.
func TargetBuilder() *Target {
	return &Target{
		err:        nil,
		infService: nil,
		exp:        nil,
		k8sclient:  nil,
	}
}

// SetK8sClient sets a k8s client within the target struct.
func (t *Target) SetK8sClient(c client.Client) target.Target {
	t.k8sclient = c
	return t
}

// SetExperiment sets a pointer to an experiment object within the target.
func (t *Target) SetExperiment(exp *experiment.Experiment) target.Target {
	t.exp = exp
	return t
}

// getNN validates components of a v1alpha2 targetRef and returns namespace and name, or error
func getNN(targetRef string) (string, string, error) {
	tc := strings.Split(targetRef, "/")
	if len(tc) == 3 && tc[0] == "v1alpha2" {
		return tc[1], tc[2], nil
	}
	return "", "", errors.New("Invalid targetRef")
}

// Fetch fetches the v1alpha2 InferenceService object fetched from the Kubernetes cluster and populates the target struct with it.
func (t *Target) Fetch(targetRef string) target.Target {
	if t.err != nil {
		return t
	}
	// figure out name and namespace of the target
	namespace, name, err := getNN(targetRef)
	if err != nil {
		t.err = errors.New("invalid target specification; v1alpha2 target needs to be of the form 'v1alpha2/inference-service-namespace/inference-service-name'")
		return t
	}
	// go get inferenceService or set an error
	t.infService = &unstructured.Unstructured{}
	t.infService.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "serving.kubeflow.org",
		Kind:    "InferenceService",
		Version: "v1alpha2",
	})
	t.err = t.k8sclient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, t.infService)
	return t
}

// Error returns the error accumulated by target until this point or nil if there is none.
func (t *Target) Error() error {
	return t.err
}

// InitializeTrafficSplit initializes traffic split for the target.
func (t *Target) InitializeTrafficSplit() target.Target {
	return t
}

// GetVersionInfo constructs the VersionInfo object based on the target and the experiment, and sets this within the target.
func (t *Target) GetVersionInfo() target.Target {
	return t
}

// GetOldBaseline gets the baseline from v1alpha2 InferenceService and sets it within the target.
func (t *Target) GetOldBaseline() target.Target {
	return t
}

// GetNewBaseline gets the recommended baseline from experiment object and sets it within the target.
func (t *Target) GetNewBaseline() target.Target {
	return t
}

// SetNewBaseline sets a new baseline (i.e., 'default' version) within the target
func (t *Target) SetNewBaseline() target.Target {
	return t
}

// SetVersionInfoInExperiment sets version info in the experiment associated with this target.
func (t *Target) SetVersionInfoInExperiment() target.Target {
	return t
}
