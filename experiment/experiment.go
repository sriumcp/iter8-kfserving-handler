// Package experiment enables extraction of useful information from experiment objects and also setting VersionInfo within them.
package experiment

import (
	"context"
	"errors"
	"os"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Experiment is an enhancement of v2alpha1.Experiment struct, and supports various methods used in describing an experiment.
type Experiment struct {
	*etc3.Experiment
}

// Builder constructs an Experiment struct with the given etc3 experiment.
func Builder(exp *etc3.Experiment) *Experiment {
	return &Experiment{exp}
}

// GetExperiment returns a pointer to the experiment object fetched from the Kubernetes cluster.
func GetExperiment(c client.Client) (*Experiment, error) {
	if name, ok := os.LookupEnv("EXPERIMENT_NAME"); ok {
		if namespace, ok := os.LookupEnv("EXPERIMENT_NAMESPACE"); ok {
			etc3Exp := &etc3.Experiment{}
			err := c.Get(context.Background(), client.ObjectKey{
				Namespace: namespace,
				Name:      name,
			}, etc3Exp)
			if err != nil {
				return nil, errors.New("Cannot get experiment: " + err.Error())
			}
			exp := &Experiment{etc3Exp}
			return exp, nil
		}
	}
	return nil, errors.New("environment variables EXPERIMENT_NAME and EXPERIMENT_NAMESPACE need to be set to valid values")
}

// GetTargetRef returns the target string for the experiment.
func (e *Experiment) GetTargetRef() string {
	return e.Spec.Target
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
		return "", errors.New("Recommended baseline not found in experiment status")
	}
	return *e.Status.RecommendedBaseline, nil
}

// GetBaseline returns the baseline version in the experiment.
func (e *Experiment) GetBaseline() (string, error) {
	if e.Spec.VersionInfo == nil {
		return "", errors.New("versionInfo not found in experiment spec")
	}
	return e.Spec.VersionInfo.Baseline.Name, nil
}

// SetVersionInfo sets version information for an experiment.
func (e *Experiment) SetVersionInfo(versionInfo *etc3.VersionInfo) {
}
