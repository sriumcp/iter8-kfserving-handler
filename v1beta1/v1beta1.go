package v1beta1

import (
	"github.com/iter8-tools/iter8-kfserving-handler/experiment"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
)

// Target is an enhancement of KFServing v1beta1 InferenceService.
type Target struct {
	err error
}

// GetTarget returns a pointer to a v1beta1 InferenceService object fetched from the Kubernetes cluster.
func GetTarget(targetRef string) *Target {
	return &Target{}
}

// Error returns the error accumulated by target until this point or nil if there is none.
func (t *Target) Error() error {
	return t.err
}

// InitializeTrafficSplit initializes traffic split for the target.
func (t *Target) InitializeTrafficSplit(exp *experiment.Experiment) error {
	return nil
}

// GetVersionInfo returns a pointer to a versionInfo struct containing the versions in the input experiment.
func (t *Target) GetVersionInfo(exp *experiment.Experiment) *etc3.VersionInfo {
	return nil
}

// SetNewBaseline sets a new baseline (i.e., 'default' version) within the target
func (t *Target) SetNewBaseline(newBaseline string) error {
	return nil
}
