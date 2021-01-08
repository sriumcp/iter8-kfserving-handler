package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTargetType(t *testing.T) {
	assert.Equal(t, V1beta1, GetTargetType("isns/isname"))
	assert.Equal(t, V1beta1, GetTargetType("v1beta1/isns/isname"))
	assert.Equal(t, V1alpha2, GetTargetType("v1alpha2/isns/isname"))
}

func TestGetCondition(t *testing.T) {
	cond := []Condition{{"Ready", "True"}, {"Finished", "False"}}
	c1, e1 := GetCondition(cond, "Ready")
	c2, e2 := GetCondition(cond, "Finished")
	_, e3 := GetCondition(cond, "Hello")
	assert.Equal(t, c1, "True")
	assert.NoError(t, e1)
	assert.Equal(t, c2, "False")
	assert.NoError(t, e2)
	assert.Error(t, e3)
}
