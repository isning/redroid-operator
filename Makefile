# Makefile for redroid-operator
#
# Prerequisites:
#   go, docker (or podman), kubectl, controller-gen, kustomize

IMG          ?= ghcr.io/isning/redroid-operator:latest
CONTROLLER_GEN ?= go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0
KUSTOMIZE    ?= kustomize
HELM         ?= helm
CHART_DIR    ?= charts/redroid-operator

##@ General

.PHONY: help
help: ## Print this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: ## Re-generate DeepCopy methods (requires controller-gen).
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manifests
manifests: ## Re-generate CRD and RBAC manifests (requires controller-gen).
	$(CONTROLLER_GEN) \
		crd:generateEmbeddedObjectMeta=true \
		rbac:roleName=redroid-operator-manager-role \
		webhook \
		paths="./..." \
		output:crd:artifacts:config=config/crd/bases \
		output:rbac:artifacts:config=config/rbac

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: test
test: ## Run unit tests (uses fake client — no cluster or envtest binary needed).
	go test ./... -v -count=1

.PHONY: test-short
test-short: ## Run tests without verbose output.
	go test ./...

.PHONY: lint
lint: ## Run golangci-lint.
	golangci-lint run ./...

.PHONY: cover
cover: ## Run tests with coverage and produce HTML report.
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -3

##@ Build

.PHONY: build
build: generate fmt vet ## Build the manager binary.
	go build -o bin/manager ./cmd/

.PHONY: build-plugin
build-plugin: ## Build the kubectl-redroid plugin binary.
	go build -ldflags "-X github.com/isning/redroid-operator/cmd/kubectl-redroid/cmd.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)" \
		-o bin/kubectl-redroid ./cmd/kubectl-redroid/

.PHONY: install-plugin
install-plugin: build-plugin ## Install kubectl-redroid to ~/.local/bin/ (must be on $$PATH).
	install -m 755 bin/kubectl-redroid ~/.local/bin/kubectl-redroid
	@echo "Installed to ~/.local/bin/kubectl-redroid"
	@echo "Make sure ~/.local/bin is in your PATH, then run: kubectl redroid --help"

.PHONY: run
run: manifests generate fmt vet ## Run against the currently configured cluster (out-of-cluster).
	go run ./cmd/

##@ Docker

.PHONY: docker-build
docker-build: ## Build the Docker image (IMG=...).
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push the Docker image.
	docker push ${IMG}

.PHONY: docker-buildx
docker-buildx: ## Build and push a multi-arch image with buildx.
	docker buildx build --platform linux/amd64,linux/arm64 -t ${IMG} --push .

##@ Deployment

.PHONY: install
install: manifests ## Install CRDs into the cluster.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: ## Remove CRDs from the cluster.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found -f -

.PHONY: deploy
deploy: manifests ## Deploy the operator to the cluster.
	cd config/manager && $(KUSTOMIZE) edit set image manager=${IMG}
	$(KUSTOMIZE) build config | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Remove the operator from the cluster.
	$(KUSTOMIZE) build config | kubectl delete --ignore-not-found -f -

##@ Helm

.PHONY: helm-crds
helm-crds: manifests ## Sync generated CRDs into the Helm chart's crds/ directory.
	cp config/crd/bases/redroid.io_redroidinstances.yaml $(CHART_DIR)/crds/
	cp config/crd/bases/redroid.io_redroidtasks.yaml     $(CHART_DIR)/crds/

.PHONY: helm-lint
helm-lint: helm-crds ## Lint the Helm chart.
	$(HELM) lint $(CHART_DIR)

.PHONY: helm-package
helm-package: helm-crds ## Package the Helm chart into dist/.
	mkdir -p dist
	$(HELM) package $(CHART_DIR) --destination dist/

##@ Helpers

.PHONY: go-sum
go-sum: ## Tidy go.mod and regenerate go.sum (run this first after cloning).
	go mod tidy

##@ Docs

CRDOC     ?= crdoc
HELM_DOCS ?= helm-docs

.PHONY: docs
docs: docs-crd docs-helm ## Generate all documentation (requires crdoc and helm-docs).

.PHONY: docs-crd
docs-crd: manifests ## Generate CRD reference docs from CRD YAML (requires crdoc).
	@command -v $(CRDOC) >/dev/null 2>&1 || \
	  { echo "crdoc not found. Install: go install fybrik.io/crdoc@latest"; exit 1; }
	mkdir -p docs/generated
	$(CRDOC) \
	  --resources config/crd/bases/ \
	  --output docs/generated/crd-reference.md \
	  --template hack/crdoc-template.tmpl 2>/dev/null || \
	$(CRDOC) \
	  --resources config/crd/bases/ \
	  --output docs/generated/crd-reference.md

.PHONY: docs-helm
docs-helm: ## Generate Helm chart README (requires helm-docs).
	@command -v $(HELM_DOCS) >/dev/null 2>&1 || \
	  { echo "helm-docs not found. Install: https://github.com/norwoodj/helm-docs"; exit 1; }
	$(HELM_DOCS) --chart-search-root charts/

