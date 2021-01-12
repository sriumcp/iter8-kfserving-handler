// Package v1beta1 provides types and methods for manipulating v1beta1 KFServing InferenceService objects.
package v1beta1

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	v1 "k8s.io/api/core/v1"
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
	retries    uint          // number of retry attempts for fetch / readiness checks
	interval   time.Duration // interval between the above attempts
}

// TargetBuilder returns an initial v1beta1 target struct pointer.
func TargetBuilder() *Target {
	return &Target{
		err:        nil,
		infService: nil,
		exp:        nil,
		k8sclient:  nil,
		retries:    18,
		interval:   10,
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
	return "", "", errors.New("Invalid targetRef")
}

// fetch is a helper function to fetch the v1beta1 InferenceService object from the Kubernetes cluster.
func (t *Target) fetch(namespace string, name string) (*unstructured.Unstructured, error) {
	isvc := &unstructured.Unstructured{}
	isvc.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "serving.kubeflow.org",
		Kind:    "InferenceService",
		Version: "v1beta1",
	})
	err := t.k8sclient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, isvc)
	return isvc, err
}

// Fetch fetches the v1beta1 InferenceService object from the Kubernetes cluster and populates the target struct with it.
// InferenceService may be unavailable at the start of this call. So, Fetch periodically attempts to fetch the InferenceService object for 180 sec.
// Upon success, it returns the fetched object; if it does not succeed in 180 secs, it returns an error.
func (t *Target) Fetch(targetRef string) target.Target {
	if t.err != nil {
		return t
	}
	// figure out name and namespace of the target
	namespace, name, err := getNN(targetRef)
	if err != nil {
		t.err = errors.New("invalid target specification; v1beta1 target needs to be of the form: 'inference-service-namespace/inference-service-name'")
		return t
	}
	// go get inferenceService or set an error
	// lines from here until return need to end up in a for loop.
	isvc, err := t.fetch(namespace, name)
	if err == nil {
		t.infService = isvc
		return t
	}
	ticker := time.NewTicker(t.interval * time.Second)
	for i := 0; i < int(t.retries); i++ {
		select {
		case <-ticker.C:
			isvc, err := t.fetch(namespace, name)
			if err == nil {
				t.infService = isvc
				return t
			}
		}
	}
	t.err = errors.New("unable to fetch target; " + err.Error())
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
	ticker := time.NewTicker(t.interval * time.Second)
	for i := 0; i < int(t.retries); i++ {
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

// SetCanaryTrafficPercent sets spec.predictor.canaryTrafficPercent field to the given value.
// After this step, the handler waits for (<=) 180 sec to ensure InferenceService object is ready.
// If any of the above steps fail, the method returns after setting an error.
func (t *Target) SetCanaryTrafficPercent(p int64) target.Target {
	if t.err != nil {
		return t
	}
	// Make sure t.infService has already been fetched.
	if t.infService == nil {
		t.err = errors.New("unable to set canary traffic split; uninitialized inference service object")
		return t
	}
	// Set spec.predictor.canaryTrafficPercent to p
	payload := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value int64  `json:"value"`
	}{{"replace", "/spec/predictor/canaryTrafficPercent", p}}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.err = errors.New("unable to marshal initial traffic split patch")
		return t
	}
	// we have made sure InferenceService object exists in the cluster, above.
	t.err = t.k8sclient.Patch(context.Background(), t.infService, client.RawPatch(types.JSONPatchType, payloadBytes))
	if t.err != nil {
		return t
	}
	r := EnsureReadiness(t)
	if !r {
		t.err = errors.New("post-patch: unable to ensure readiness of inference service even after 180 seconds")
	}
	return t
}

// InitializeTrafficSplit initializes traffic split for the target.
// The value of the field spec.predictor.canaryTrafficPercent is set to 1 (i.e., 1%).
// After this step, the handler waits for (<=) 180 sec to ensure InferenceService object is ready.
// If any of the above steps fail, the method returns after setting an error.
func (t *Target) InitializeTrafficSplit() target.Target {
	return t.SetCanaryTrafficPercent(1)
}

// GetVersionInfo constructs the VersionInfo object based on the target and returns it.
func (t *Target) GetVersionInfo() (*etc3.VersionInfo, error) {
	// candidate
	cRev, b1, err1 := unstructured.NestedString(t.infService.Object, "status", "components", "predictor", "latestCreatedRevision")
	// baseline
	bRev, b2, err2 := unstructured.NestedString(t.infService.Object, "status", "components", "predictor", "latestRolledoutRevision")

	if b1 == false || b2 == false || err1 != nil || err2 != nil {
		return nil, errors.New("unable to extract default and canary revisions from target")
	}

	ns, name, err3 := getNN(t.exp.GetTargetRef())
	if err3 != nil {
		return nil, errors.New("unable to extract name and namespace of target")
	}

	vi := etc3.VersionInfo{
		Baseline: etc3.VersionDetail{
			Name: "default",
			Tags: &map[string]string{"revision": bRev},
		},
		Candidates: []etc3.VersionDetail{
			{
				Name: "canary",
				Tags: &map[string]string{"revision": cRev},
				WeightObjRef: &v1.ObjectReference{
					Kind:       "InferenceService",
					Namespace:  ns,
					Name:       name,
					APIVersion: "serving.kubeflow.org/v1beta1",
					FieldPath:  "/spec/predictor/canaryTrafficPercent",
				},
			},
		},
	}
	return &vi, nil
}

// SetVersionInfoInExperiment sets version info in the experiment associated with this target.
func (t *Target) SetVersionInfoInExperiment() target.Target {
	if t.err != nil {
		return t
	}
	// get versionInfo
	var vi *etc3.VersionInfo
	vi, t.err = t.GetVersionInfo()
	if t.err != nil {
		return t
	}
	if vi == nil {
		t.err = errors.New("Could not get versionInfo for experiment")
		return t
	}
	if !t.exp.IsSingleVersion() && len(vi.Candidates) == 0 {
		t.err = errors.New("expected baseline and candidate; did not find candidate during GetVersionInfo")
		return t
	}
	// patch experiment with versionInfo
	payload := []struct {
		Op    string            `json:"op"`
		Path  string            `json:"path"`
		Value *etc3.VersionInfo `json:"value"`
	}{{"replace", "/spec/versionInfo", vi}}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.err = errors.New("unable to marshal experiment versionInfo patch")
		return t
	}
	t.err = t.k8sclient.Patch(context.Background(), t.exp.Experiment, client.RawPatch(types.JSONPatchType, payloadBytes))
	return t
}

// SetNewBaseline sets a new baseline (i.e., 'default' version) within the target
func (t *Target) SetNewBaseline() target.Target {
	if t.err != nil {
		return t
	}
	if t.exp == nil {
		t.err = errors.New("method SetNewBaseline called on a target with nil experiment")
		return t
	}
	if t.exp.IsSingleVersion() {
		t.err = errors.New("method SetNewBaseline called on a target with a single-version experiment")
		return t
	}
	recommendedBaseline, err := t.exp.GetRecommendedBaseline()
	if err != nil {
		t.err = errors.New("error in getting recommended baseline from experiment")
		return t
	}
	if recommendedBaseline == "canary" {
		return t.SetCanaryTrafficPercent(100)
	}
	return t.SetCanaryTrafficPercent(0)
}
