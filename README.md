# Knative Sample Controller

[![GoDoc](https://godoc.org/knative.dev/sample-controller?status.svg)](https://godoc.org/knative.dev/sample-controller)
[![Go Report Card](https://goreportcard.com/badge/knative/sample-controller)](https://goreportcard.com/report/knative/sample-controller)

Knative `servinng events controller` defines a controller for `knative serving` resource and listens for Knative Service events and generates corresponding `CloudEvents` taking the [tektoncd controller](https://github.com/tektoncd/experimental/tree/main/cloudevents) as a reference implementation.

### Controller architecture

The following roughly defines the controller components:
1. We register 2 Informers:
    * [ServingV1/Service](knative.dev/serving/pkg/client/injection/informers/serving/v1/service)
    * [ServiceV1/Revision](knative.dev/serving/pkg/client/injection/informers/serving/v1/revision)
2. A `ServingV1/Service` reconciler
3. Embed the [CloudEvents](github.com/cloudevents/sdk-go/v2) SDK client.
4. We also register a CloudEvents 'receiver' which listens at port `8080`.

### Testing and running the controller
1. Setup a test cluster using KIND with `make cluster`
2. Install Knative with `make install-knative`
3. Install CRDs with `make install-crds`
4. Run the controller with `make run-controllers`

#### Scenario 1 - Serving object to CloudEvent
Now create a dummy Knative service 
```
echo "apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: hello
spec:
  template:
    spec:
      containers:
      - image: gcr.io/google-samples/hello-app:1.0
" | kubectl apply -f -
```

We will see the corresponding events received by the controller and subsequently the emitted `cloudevent` received:
```
Context Attributes,
  specversion: 1.0
  type: cd.service.deployed.v1      # event of type deployed
  source: default/hello
  id: 1fc4b160-e859-4177-8847-b85c19fdbc18
  time: 2022-01-20T16:56:39.010077Z
```

On updating the previously created `ksvc` object, we should see a corresponding event of type `upgraded`
```diff
- - image: gcr.io/google-samples/hello-app:1.0
+ - image: gcr.io/google-samples/hello-app:2.0
```
This generates the upgraded event and should be reflected in the event as well:
```
Context Attributes,
  specversion: 1.0
  type: cd.service.upgraded.v1      # event of type upgraded
  source: default/hello
  id: 23225ed1-3dcc-4772-a60d-72ab57bb1ba3
  time: 2022-01-20T17:16:00.253051Z

```

#### Scenario 2 - CloudEvent triggering a Knative service creation
1. Clearing the previous `ksvc` object
```
kubectl delete ksvc hello
```
2. Trigger a cloud event of type `created` (`cd.service.created.v1`) with the following command:
```
go run cmd/cloudevent/produce.go
```
This should create a `ksvc` object in the cluster
> note that the event used here (`cd.service.created.v1`) is not yet defined by the CloudEvents sdk and is just hardcoded for demo purposes.

Check for the created Knative service by running:
```
kubectl get ksvc 
```

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
