module github.com/signalfx/splunk-otel-collector-operator

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef // indirect
	golang.org/x/tools v0.1.7 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/kubectl v0.22.2
	sigs.k8s.io/controller-runtime v0.9.6
)
