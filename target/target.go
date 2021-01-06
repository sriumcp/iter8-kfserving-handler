package target

import (
	"strings"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
)

// Target interface represents the target of an iter8 experiment.
type Target interface {
	Error() error
	InitializeTrafficSplit(exp *experiment.Experiment) error
	GetVersionInfo(exp *experiment.Experiment) *etc3.VersionInfo
	SetNewBaseline(newBaseline string) error
}

// ISType is the type of InferenceService object. In KFServing, this may be v1beta1 or v1alpha2.
type ISType string

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
