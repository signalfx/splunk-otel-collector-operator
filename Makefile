VERSION ?= "$(shell grep -v '\#' versions.txt | grep operator | awk -F= '{print $$2}')"
VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION_PKG = "github.com/signalfx/splunk-otel-collector-operator/internal/version"
VERSION_COLLECTOR ?= "$(shell grep -v '\#' versions.txt | grep splunk-otel-collector | awk -F= '{print $$2}')"
LD_FLAGS ?= "-X ${VERSION_PKG}.version=${VERSION} -X ${VERSION_PKG}.buildDate=${VERSION_DATE} -X ${VERSION_PKG}.collectorVersion=${VERSION_COLLECTOR}"

# Image URL to use all building/pushing image targets
QUAY_USER ?= signalfx
IMG_PREFIX ?= quay.io/${QUAY_USER}
IMG_REPO ?= splunk-otel-operator
IMG ?= ${IMG_PREFIX}/${IMG_REPO}:$(addprefix v,${VERSION})

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

OPERATOR_SDK=$(shell which operator-sdk)
KUTTL=$(shell which kubectl-kuttl)
CONTROLLER_GEN = $(shell which controller-gen)
KUSTOMIZE = $(shell which kustomize)

KUBE_VERSION ?= 1.24
KIND_CONFIG ?= kind-$(KUBE_VERSION).yaml

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

# ensure-generate-is-noop: VERSION=$(OPERATOR_VERSION)
.PHONY: ensure-generate-is-noop
ensure-generate-is-noop: USER=signalfx
ensure-generate-is-noop: set-image-controller generate bundle
	if [[ `git status --porcelain` ]]; then \
		git diff; \
		echo "Build failed: a model has been changed but the generated resources aren't up to date. Run 'make generate' and update your PR." && exit 1; \
	else \
		echo "All models are in sync with generated resources."; \
	fi

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases 

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="./hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run --allow-parallel-runners

.PHONY: lint-all
lint-all: generate ensure-generate-is-noop fmt vet lint

ENVTEST=$(shell pwd)/testbin
#test: manifests generate fmt vet ## Run tests.
.PHONY: test
test:
	mkdir -p ${ENVTEST}
	# setup-envtest creates the api-server, etcd and kubectl binaries in the ENVTEST/KUBEBUILDER_ASSETS directory.
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(KUBE_VERSION) -p path --bin-dir $(ENVTEST))" go test ${GOTEST_OPTS} ./...

##@ Build
.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager -ldflags ${LD_FLAGS} main.go
.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run -ldflags=${LD_FLAGS} ./main.go
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build --build-arg VERSION_DATE=${VERSION_DATE} --build-arg VERSION_PKG=${VERSION_PKG} --build-arg VERSION_COLLECTOR=${VERSION_COLLECTOR} --build-arg VERSION=${VERSION} -t ${IMG} .
.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment
.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -
.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -
.PHONY: deploy
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -
.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -
.PHONY: install-tools
install-tools : ## Download CLI tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest # Only used to setup testing environments
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
	go install sigs.k8s.io/kustomize/kustomize/v4

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
define go-get-tool-old
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Set the controller image parameters
.PHONY: set-image-controller
set-image-controller: manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG) 
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --manifests --metadata --version $(VERSION)
	$(OPERATOR_SDK) bundle validate ./bundle

# dirty hack for now
.PHONY: bundle-openshift
bundle-openshift: bundle
	rm -rf bundle-openshift
	cp -r bundle bundle-openshift
	cat config/openshift/*.yaml >> bundle-openshift/manifests/splunk-otel-operator-role_rbac.authorization.k8s.io_v1_clusterrole.yaml

# Generates the released manifests
.PHONY: release-artifacts
release-artifacts: set-image-controller
	mkdir -p dist
	$(KUSTOMIZE) build config/default -o dist/splunk-otel-operator.yaml
	# dirty hack for now
	cp dist/splunk-otel-operator.yaml dist/splunk-otel-operator-openshift.yaml
	cat config/openshift/*.yaml >> dist/splunk-otel-operator-openshift.yaml

##@ Tests
.PHONY: e2e
e2e: ## Run end-to-tests
	$(KUTTL) test
.PHONY: prepare-e2e
prepare-e2e: set-test-image-vars set-image-controller docker-build start-kind ## prepare end-to-end tests
	mkdir -p tests/_build/crds tests/_build/manifests
	$(KUSTOMIZE) build config/default -o tests/_build/manifests/01-splunk-otel-operator.yaml
	$(KUSTOMIZE) build config/crd -o tests/_build/crds/

.PHONY: clean-e2e
clean-e2e: ## delete kind cluster
	kind delete cluster

.PHONY: set-test-image-vars
set-test-image-vars:
	$(eval IMG=local/splunk-otel-operator:e2e)

.PHONY: start-kind
start-kind:
	if kind get clusters | grep kind; then \
		echo "kind cluster has already been created"; \
	else \
		kind create cluster --config $(KIND_CONFIG); \
	fi

	kind load docker-image local/splunk-otel-operator:e2e

.PHONY: cert-manager
cert-manager:
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.5.2/cert-manager.yaml
	kubectl wait --timeout=5m --for=condition=available deployment cert-manager -n cert-manager
	kubectl wait --timeout=5m --for=condition=available deployment cert-manager-cainjector -n cert-manager
	kubectl wait --timeout=5m --for=condition=available deployment cert-manager-webhook -n cert-manager
