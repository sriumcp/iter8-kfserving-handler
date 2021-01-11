// Package target provides types and methods for targets of iter8 experiments.
package target

import (
	"errors"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Target interface represents the target of an iter8 experiment.
type Target interface {
	Error() error
	InitializeTrafficSplit() Target
	GetVersionInfo() (*etc3.VersionInfo, error)
	SetNewBaseline() Target
	SetExperiment(exp *experiment.Experiment) Target
	SetK8sClient(c client.Client) Target
	Fetch(targetRef string) Target
	SetVersionInfoInExperiment() Target
}

// PatchInt64Value specifies the patch data needed to patch a int64 field.
type PatchInt64Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int64  `json:"value"`
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
