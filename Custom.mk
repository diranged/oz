SOURCE := $(wildcard api/*/*.go controller/*.go ozctl/*.go ozctl/*/*.go)

## Tool Binaries
HELMIFY_VER ?= v0.3.18
HELMIFY     ?= $(LOCALBIN)/helmify

HELM_DOCS_VER ?= v1.11.0
HELM_DOCS     ?= $(LOCALBIN)/helm-docs

REVIVE_VER ?= v1.2.4
REVIVE     ?= $(LOCALBIN)/revive

.PHONY: docker-load
docker-load:
	kind load docker-image $(IMG) -n $(KIND_CLUSTER_NAME)

.PHONY: cover
cover:
	go tool cover -func cover.out

.PHONY: lint
lint: revive
	$(REVIVE) -config revive.toml -formatter stylish ./...

.PHONY: test-e2e  # you will need to have a Kind cluster up and running to run this target
test-e2e:
	go test ./test/e2e/ -v -ginkgo.v


.PHONY: revive
revive: $(REVIVE) ## Download revive locally if necessary.
$(REVIVE): $(LOCALBIN) Custom.mk
	GOBIN=$(LOCALBIN) go install github.com/mgechev/revive@$(REVIVE_VER)

##@ Build CLI
.PHONY: cli
cli: outputs/ozctl-osx outputs/ozctl-osx-arm64

outputs/ozctl-osx: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=amd64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl

outputs/ozctl-osx-arm64: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=arm64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl

## https://github.com/arttor/helmify#integrate-to-your-operator-sdkkubebuilder-project
$(HELMIFY): $(LOCALBIN) Custom.mk
	GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@$(HELMIFY_VER)

helm: manifests kustomize $(HELMIFY)
	$(KUSTOMIZE) build config/default | $(HELMIFY) \
		-crd-dir \
		charts/oz

$(HELM_DOCS): $(LOCALBIN) Custom.mk
	GO111MODULE=on GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VER)

.PHONY: helm-docs
helm-docs: $(HELM_DOCS)
	$(HELM_DOCS)
	git diff --exit-code
