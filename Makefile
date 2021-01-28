
# Image URL to use all building/pushing image targets
IMG ?= terraform-k8s:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

KUSTOMIZE=$(shell which kustomize)
ifeq ($(.SHELLSTATUS),1)
	$(error "kustomize not found. Please follow the instructions here to install it: https://kubectl.docs.kubernetes.io/installation/kustomize/")
endif
CONTROLLER_GEN=$(shell which controller-gen)
ifeq ($(.SHELLSTATUS),1)
	$(error "controller-gen not found. Please install by running: go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0")
endif
KUBEBUILDER=$(shell which kubebuilder)
ifeq ($(.SHELLSTATUS),1)
	$(error "Kubebuilder and related assets such as the etcd binary could not be found in PATH. Please install kubebuilder as explained here: https://book.kubebuilder.io/quick-start.html#installation")
endif
export KUBEBUILDER_ASSETS ?= $(dir $(KUBEBUILDER))

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: test deploy

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/terraform-k8s main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=terraform-k8s webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate:
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
