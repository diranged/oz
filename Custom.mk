SOURCE := $(wildcard api/*/*.go controller/*.go ozctl/*.go ozctl/*/*.go)
REVIVE_VER := v1.2.4

## Tool Binaries
REVIVE ?= $(LOCALBIN)/revive

.PHONY: docker-load
docker-load:
	kind load docker-image $(IMG) -n $(KIND_CLUSTER_NAME)

.PHONY: cover
cover:
	go tool cover -func cover.out

.PHONY: lint
lint:
	$(REVIVE) -set_exit_status -formatter stylish ./...

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
