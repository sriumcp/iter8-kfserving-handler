package v1alpha2

import (
	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Target is an enhancement of KFServing v1alpha2 InferenceService.
type Target struct {
	err        error
	infService *unstructured.Unstructured
	exp        *experiment.Experiment
}

// GetTarget fetches the v1alpha2 InferenceService object fetched from the Kubernetes cluster and populates the target struct with it.
func (t *Target) GetTarget() target.Target {
	// go get inferenceService or set an error
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

// SetExperiment sets a pointer to an experiment object within the target.
func (t *Target) SetExperiment(exp *experiment.Experiment) target.Target {
	return t
}

// TargetBuilder returns an initial v1beta1 target struct pointer.
func TargetBuilder() *Target {
	return &Target{
		err: nil,
		exp: nil,
	}
}

// SetVersionInfoInExperiment sets version info in the experiment associated with this target.
func (t *Target) SetVersionInfoInExperiment() target.Target {
	return t
}
