SOURCE := $(wildcard api/*/*.go controller/*.go ozctl/*.go ozctl/*/*.go)
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le

## Tool Binaries
REVIVE_VER ?= v1.2.4
REVIVE     ?= $(LOCALBIN)/revive

.PHONY: docker-load
docker-load:
	kind load docker-image $(IMG) -n $(KIND_CLUSTER_NAME)

# override the makefile docker-buildx which is broken, and use a simpler one anyways
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- DOCKER_BUILDKIT=1 docker buildx build \
		$(BUILD_ARGS) \
		--cache-from type=local,src=.buildx_cache \
		--cache-to type=local,dest=.buildx_cache \
		--platform=$(PLATFORMS) \
		--tag ${IMG} .
	- docker buildx rm project-v3-builder

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
$(REVIVE): $(LOCALBIN)
	test -s $(LOCALBIN)/revive || GOBIN=$(LOCALBIN) go install github.com/mgechev/revive@$(REVIVE_VER)

##@ Build CLI
.PHONY: cli
cli: outputs/ozctl-osx outputs/ozctl-osx-arm64

outputs/ozctl-osx: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=amd64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl

outputs/ozctl-osx-arm64: ozctl controllers api $(SOURCE)
	GOOS=darwin GOARCH=arm64 LDFLAGS=$(RELEASE_LDFLAGS) go build -o $@ ./ozctl
