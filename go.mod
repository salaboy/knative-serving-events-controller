module github.com/salaboy/knative-serving-events-controller

go 1.15

require (
	github.com/cdfoundation/sig-events/cde/sdk/go v0.0.0-20211122192319-2ad36f58fc2c
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/satori/go.uuid v1.2.0
	google.golang.org/api v0.62.0 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	k8s.io/code-generator v0.22.5
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
	knative.dev/hack v0.0.0-20211222071919-abd085fc43de
	knative.dev/pkg v0.0.0-20220104185830-52e42b760b54
	knative.dev/serving v0.28.0
)
