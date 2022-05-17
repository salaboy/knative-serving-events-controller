.EXPORT_ALL_VARIABLES:
SYSTEM_NAMESPACE ?= default
METRICS_DOMAIN ?= example.com
CLUSTER_NAME ?= knative-test
EVENTSINK ?= https://broker.ishankhare.dev/default/default
KO_DOCKER_REPO ?= ishankhare07
# KIND_CLUSTER_NAME ?= knative-test
CRED_PATH ?= /Users/ishankhare/Downloads/tonal-baton-181908-c09004360185.json

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
	# go run cmd/controller/main.go
	ko resolve -RBf config/ko # builds and pushes to docker registry

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

install-crossplane:
	kubectl create namespace crossplane-system
	helm repo add crossplane-stable https://charts.crossplane.io/stable
	helm repo update
	helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
	sleep 40
	kubectl apply -f config/crossplane/gcp/provider.yaml
	sleep 15
	kubectl apply -f config/crossplane/gcp/provider-config.yaml
	kubectl create secret generic gcp-creds -n crossplane-system --from-file=creds=${CRED_PATH}

create-workload-cluster:
	kubectl apply -f config/crossplane/resources/
	echo "setting up helm provider"
	kubectl apply -f config/crossplane/helm/provider.yaml
	sleep 15
	kubectl apply -f config/crossplane/helm/provider-config.yaml
	kubectl apply -f config/crossplane/helm/release.yaml

setup-tekton:
	kubectl apply -f tekton/resources/

install: install-knative install-tekton setup-tekton install-crossplane
