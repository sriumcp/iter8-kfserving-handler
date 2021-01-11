module github.com/iter8-tools/iter8-kfserving-handler

go 1.15

require (
	github.com/iter8-tools/etc3 v0.1.0-rc
	github.com/iter8-tools/iter8ctl v0.0.0-20210106155027-e6d413ca9b02
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.0 // indirect
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/yaml v1.2.0
)
