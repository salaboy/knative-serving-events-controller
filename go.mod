module knative.dev/sample-controller

go 1.15

require (
	github.com/cdfoundation/sig-events/cde/sdk/go v0.0.0-20211122192319-2ad36f58fc2c
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/satori/go.uuid v1.2.0
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	k8s.io/code-generator v0.22.5
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
	knative.dev/hack v0.0.0-20211222071919-abd085fc43de
	knative.dev/pkg v0.0.0-20220104185830-52e42b760b54
	knative.dev/serving v0.28.0
)
