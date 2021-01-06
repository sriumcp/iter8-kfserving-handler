package experiment

import (
	"errors"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
)

// Experiment is an enhancement of v2alpha1.Experiment struct, and supports various methods used in describing an experiment.
type Experiment struct {
	etc3.Experiment
	err error
}

// GetExperiment returns a pointer to the experiment object fetched from the Kubernetes cluster.
func GetExperiment() (*Experiment, error) {
	return nil, nil
}

// GetTargetRef returns the target string for the experiment.
func (e *Experiment) GetTargetRef() string {
	return e.Spec.Target
}

// SetVersionInfo sets version information for an experiment.
func (e *Experiment) SetVersionInfo(versionInfo *etc3.VersionInfo) {
}

// Error returns the error accumulated by experiment until this point or nil if there is none.
func (e *Experiment) Error() error {
	return e.err
}

// IsSingleVersion returns a boolean indicating if this is a single version experiment.
func (e *Experiment) IsSingleVersion() bool {
	if e.Spec.Strategy.Type == etc3.StrategyTypePerformance {
		return true
	}
	return false
}

// GetRecommendedBaseline returns the next baseline recommended in the experiment.
func (e *Experiment) GetRecommendedBaseline() (string, error) {
	if e.Status.RecommendedBaseline == nil {
		e.err = errors.New("Recommended baseline not found in experiment status")
		return "", e.err
	}
	return *e.Status.RecommendedBaseline, nil
}

// GetBaseline returns the baseline version in the experiment.
func (e *Experiment) GetBaseline() (string, error) {
	if e.Spec.VersionInfo == nil {
		e.err = errors.New("versionInfo not found in experiment spec")
		return "", e.err
	}
	return e.Spec.VersionInfo.Baseline.Name, nil
}
