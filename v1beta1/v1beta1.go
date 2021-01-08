package v1beta1

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Target is an enhancement of KFServing v1beta1 InferenceService.
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

// Error returns the error accumulated by target until this point or nil if there is none.
func (t *Target) Error() error {
	return t.err
}

// SetK8sClient sets a k8s client within the target struct.
func (t *Target) SetK8sClient(c client.Client) target.Target {
	if t.err != nil {
		return t
	}
	t.k8sclient = c
	return t
}

// SetExperiment sets a pointer to an experiment object within the target.
func (t *Target) SetExperiment(exp *experiment.Experiment) target.Target {
	if t.err != nil {
		return t
	}
	t.exp = exp
	return t
}

// getNN validates components of a v1beta1 targetRef and returns namespace and name, or error
func getNN(targetRef string) (string, string, error) {
	tc := strings.Split(targetRef, "/")
	if len(tc) == 2 {
		return tc[0], tc[1], nil
	}
	if len(tc) == 3 {
		if tc[0] == "v1beta1" {
			return tc[1], tc[2], nil
		}
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
		t.err = errors.New("invalid target specification; v1beta1 target needs to be in one of two forms: 'inference-service-namespace/inference-service-name' or 'v1beta1/inference-service-namespace/inference-service-name'")
		return t
	}
	// go get inferenceService or set an error
	t.infService = &unstructured.Unstructured{}
	t.infService.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "serving.kubeflow.org",
		Kind:    "InferenceService",
		Version: "v1beta1",
	})
	t.err = t.k8sclient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, t.infService)
	return t
}

// getCond is a helper function for fetching the target and getting its readiness.
func getCond(t *Target) bool {
	t.Fetch(t.exp.GetTargetRef())
	cond, err := GetConditions(t)
	if err == nil {
		readyStr, _ := target.GetCondition(cond, "Ready")
		return (readyStr == "True")
	}
	return false
}

// EnsureReadiness ensures that the condition "Ready" has "Status" true in t.infService.
// It periodically fetches t.infService and checks this condition.
// Returns true if readiness is reached in 180 sec and false otherwise.
func EnsureReadiness(t *Target) bool {
	ready := getCond(t)
	if ready {
		return true
	}
	ticker := time.NewTicker(10 * time.Second)
	for i := 0; i < 18; i++ {
		select {
		case <-ticker.C:
			ready = getCond(t)
			if ready {
				return true
			}
		}
	}
	return false
}

// InitializeTrafficSplit initializes traffic split for the target.
// The value of the field spec.predictor.canaryTrafficPercent is set to 1 (i.e., 1%).
// After this step, the handler waits for (<=) 200 sec to ensure InferenceService object is ready.
// If any of the above steps fail, the handler exits with an error.
func (t *Target) InitializeTrafficSplit() target.Target {
	if t.err != nil {
		return t
	}
	// Make sure t.infService has already been fetched.
	if t.infService == nil {
		t.err = errors.New("unable to initialize traffic split; uninitialize inference service object")
		return t
	}
	// Set spec.predictor.canaryTrafficPercent to 1%
	payload := []target.PatchInt64Value{{
		Op:    "replace",
		Path:  "/spec/predictor/canaryTrafficPercent",
		Value: 1,
	}}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.err = errors.New("unable to marshal initial traffic split patch")
		return t
	}
	t.err = t.k8sclient.Patch(context.Background(), t.infService, client.RawPatch(types.JSONPatchType, payloadBytes))

	// There needs to be a readiness check here...
	r := EnsureReadiness(t)
	if !r {
		t.err = errors.New("unable to ensure readiness of inference service even after 180 seconds")
	}
	return t
}

// GetVersionInfo constructs the VersionInfo object based on the target and the experiment, and sets this within the target.
func (t *Target) GetVersionInfo() target.Target {
	return t
}

// GetConditions unmarshals conditions from status and returns a slice of conditions.
func GetConditions(t *Target) ([]target.Condition, error) {
	if t.err != nil {
		return nil, errors.New("GetConditions called on erroneous target")
	}
	type resource struct {
		Status struct {
			Conditions []target.Condition `json:"conditions"`
		} `json:"status"`
	}
	var ro = resource{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(t.infService.Object, &ro)
	return ro.Status.Conditions, err
}

// GetOldBaseline gets the baseline from v1beta1 InferenceService and sets it within the target.
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
