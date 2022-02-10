.EXPORT_ALL_VARIABLES:
SYSTEM_NAMESPACE ?= default
METRICS_DOMAIN ?= example.com
CLUSTER_NAME ?= knative-test
EVENTSINK ?= http://localhost:8080
KO_DOCKER_REPO ?= localhost:5000
KIND_CLUSTER_NAME ?= knative-test

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
	for crd in config/crds; \
	do \
		kubectl apply -f "$$crd"; \
	done

run-controllers:
	go run cmd/controller/main.go
	# ko resolve -Rf config/ko

run-webhooks:
	go run cmd/webhook/main.go

run-schema:
	go run cmd/schema/main.go dump SimpleDeployment

cluster:
	kind create cluster --name ${CLUSTER_NAME} --config kind-config.yaml

delete-cluster:
	kind delete cluster --name ${CLUSTER_NAME}
	docker stop kind-registry
	docker rm kind-registry

cluster-with-registry:
	./kind-with-registry.sh

install-tekton:
	echo "installing tekton pipelines"
	kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
	echo "installing tekton triggers"
	kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
	kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
	echo "applying tekton user, role and rolebinding"
	kubectl apply -f tekton/rbac/admin-role.yaml
	kubectl apply -f tekton/rbac/crb.yaml 
	kubectl apply -f tekton/rbac/trigger-webhook-role.yaml

setup-tekton:
	kubectl apply -f tekton/resources/
