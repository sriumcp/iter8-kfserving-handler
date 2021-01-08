package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIter8KfservingHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iter8KfservingHandler Suite")
}
