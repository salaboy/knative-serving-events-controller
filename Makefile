.EXPORT_ALL_VARIABLES:
SYSTEM_NAMESPACE ?= default
METRICS_DOMAIN ?= example.com
CLUSTER_NAME ?= knative-test

install-knative:
	kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.1.0/serving-crds.yaml
	kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.1.0/serving-core.yaml
	# install networking layer | kourier
	kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.1.0/kourier.yaml
	kubectl patch configmap/config-network \
		  --namespace knative-serving \
		    --type merge \
			  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'


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

cluster:
	kind create cluster --name ${CLUSTER_NAME} --config kind-config.yaml

delete-cluster:
	kind delete cluster --name ${CLUSTER_NAME}
