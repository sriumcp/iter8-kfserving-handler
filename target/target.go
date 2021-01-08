package target

import (
	"errors"
	"strings"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Target interface represents the target of an iter8 experiment.
type Target interface {
	Error() error
	InitializeTrafficSplit() Target
	GetVersionInfo() Target
	GetOldBaseline() Target
	GetNewBaseline() Target
	SetNewBaseline() Target
	SetExperiment(exp *experiment.Experiment) Target
	SetK8sClient(c client.Client) Target
	Fetch(targetRef string) Target
	SetVersionInfoInExperiment() Target
}

// ISType is the type of InferenceService object. In KFServing, this may be v1beta1 or v1alpha2.
type ISType string

// PatchInt64Value specifies the patch data needed to patch a int64 field.
type PatchInt64Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int64  `json:"value"`
}

const (
	//V1beta1 refers to KFServing InferenceService V1beta1 API
	V1beta1 ISType = "v1beta1"
	//V1alpha2 refers to KFServing InferenceService V1alpha2 API
	V1alpha2 ISType = "v1alpha2"
)

// GetTargetType returns the type of target based on the target string. If target string is of the form v1alpha2/namespace/name, then v1alpha2 is the returned type. Otherwise, v1beta1 is the returned type.
func GetTargetType(targetRef string) ISType {
	if strings.HasPrefix(targetRef, string(V1alpha2)) {
		return V1alpha2
	}
	return V1beta1
}

// Condition defines a readiness condition for InferenceService
type Condition struct {
	// Type of condition.
	Type string `json:"type" description:"type of status condition"`

	// Status of the condition, one of True, False, Unknown.
	Status string `json:"status" description:"status of the condition, one of True, False, Unknown"`
}

// GetCondition returns the status of a given condition by checking a slice of conditions.
func GetCondition(cond []Condition, ctype string) (string, error) {
	for _, c := range cond {
		if c.Type == ctype {
			return c.Status, nil
		}
	}
	return "", errors.New("Non existing condition")
}
