SOURCE := $(wildcard api/*/*.go controller/*.go ozctl/*.go ozctl/*/*.go)

## Tool Binaries
REVIVE_VER ?= v1.2.4
REVIVE     ?= $(LOCALBIN)/revive

GOFUMPT_VER ?= v0.4.0
GOFUMPT     ?= $(LOCALBIN)/gofumpt

GOLINES_VER ?= v0.11.0
GOLINES     ?= $(LOCALBIN)/golines

GEN_CRD_API_DOCS_VER ?= v0.3.1-0.20220223025230-af7c5e0048a3
GEN_CRD_API_DOCS     ?= $(LOCALBIN)/go-crd-api-reference-docs

.PHONY: docker-load
docker-load:
	kind load docker-image $(IMG) -n $(KIND_CLUSTER_NAME)

.PHONY: cover
cover:
	go tool cover -func cover.out

.PHONY: coverhtml
coverhtml:
	go tool cover -html cover.out

.PHONY: lint
lint: revive
	$(REVIVE) -config revive.toml -formatter stylish ./...

.PHONY: test-e2e  # you will need to have a Kind cluster up and running to run this target
test-e2e:
	go test ./test/e2e/ -v -ginkgo.v

.PHONY: gofumpt
gofumpt: $(GOFUMPT)
$(GOFUMPT):
	GOBIN=$(LOCALBIN) go install mvdan.cc/gofumpt@$(GOFUMPT_VER)

.PHONY: golines
golines: $(GOLINES)
$(GOLINES):
	GOBIN=$(LOCALBIN) go install github.com/segmentio/golines@$(GOLINES_VER)

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GOFUMPT) -l -w .
	$(GOLINES) -w .

.PHONY: revive
revive: $(REVIVE) ## Download revive locally if necessary.
$(REVIVE): $(LOCALBIN) Custom.mk
	GOBIN=$(LOCALBIN) go install github.com/mgechev/revive@$(REVIVE_VER)

gen-crd-api-reference-docs: $(GEN_CRD_API_DOCS)
$(GEN_CRD_API_DOCS):
	GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_DOCS_VER)

.PHONY: godocs
godocs: $(GEN_CRD_API_DOCS)
	bin/gen-crd-api-reference-docs \
		-config ./gen-crd-api-reference-docs.json \
		-api-dir ./api/v1alpha1 \
		-template-dir $$(go env GOMODCACHE)/github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_DOCS_VER)/template \
		-out-file docs.md \
		-v 5

##@ Build CLI
.PHONY: cli
cli: outputs/ozctl-osx outputs/ozctl-osx-arm64

outputs/ozctl-osx: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=amd64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl

outputs/ozctl-osx-arm64: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=arm64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl
