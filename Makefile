.EXPORT_ALL_VARIABLES:
SYSTEM_NAMESPACE ?= default
METRICS_DOMAIN ?= example.com

install-knative-operator:
	kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.1.0/operator.yaml

install-knative-serving:
	kubectl apply -f knative-serving.yaml

install-crds:
	for crd in config; \
	do \
		kubectl apply -f "$$crd"; \
	done

run-controllers:
	go run cmd/controller/main.go

run-webhooks:
	go run cmd/webhook/main.go

run-schema:
	go run cmd/schema/main.go dump SimpleDeployment
