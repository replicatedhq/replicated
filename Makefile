API_PKGS=apps channels releases

BUILDTAGS = containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp

export GO111MODULE=on

export CGO_ENABLED=0

.PHONY: test-unit
test-unit:
	go test -v `go list ./... | grep -v /pact | grep -v /pkg/integration` -tags "$(BUILDTAGS)"

.PHONY: test-pact
test-pact:
	go test -v ./pact/... -tags "$(BUILDTAGS)"

.PHONY: test-integration
test-integration: build
	go test -v ./pkg/integration/...

.PHONY: test-lint
test-lint: build
	./scripts/test-lint.sh

.PHONY: publish-pact
publish-pact:
	pact-broker publish ./pacts \
		--auto-detect-version-properties \
		--consumer-app-version ${PACT_VERSION} \
		--verbose
	$(MAKE) unpublish-past-versions

.PHONY: can-i-deploy
can-i-deploy:
	pact-broker can-i-deploy \
		--pacticipant replicated-cli \
		--version ${PACT_VERSION} \
		--to-environment production \
		--verbose

.PHONY: unpublish-past-versions
unpublish-past-versions:
	./scripts/unpublish-past-versions.sh

# fetch the swagger specs from the production Vendor API
.PHONY: get-spec-prod
get-spec-prod:
	mkdir -p gen/spec/
	curl -o gen/spec/v1.json https://api.replicated.com/vendor/v1/spec/vendor-api.json
	curl -o gen/spec/v3.json https://api.replicated.com/vendor/v3/spec/vendor-api-v3.json

# generate the swagger specs from the local replicatedcom/vendor-api repo
.PHONY: get-spec-local
get-spec-local:
	mkdir -p gen/spec/
	docker run --rm \
		--volume ${GOPATH}/src/github.com:/go/src/github.com \
		replicatedhq.replicated /bin/bash -c ' \
			for PKG in ${API_PKGS}; do \
				swagger generate spec \
					-b ../../replicatedcom/vendor-api/handlers/replv1/$$PKG \
					-o gen/spec/$$PKG.json; \
			done \
			&& swagger generate spec \
				-b ../../replicatedcom/vendor-api/handlers/replv2 \
				-o gen/spec/v2.json'

# generate from the specs in gen/spec, which come from either get-spec-prod or get-spec-local
.PHONY: gen-models
gen-models:
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i /local/gen/spec/v1.json \
		-l go \
		-o /local/gen/go/v1; \
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i /local/gen/spec/v3.json \
		-l go \
		-o /local/gen/go/v3; \

.PHONY: build
build:
	go build \
		${LDFLAGS} \
		-tags "$(BUILDTAGS)" \
		-o bin/replicated \
		cli/main.go

.PHONY: release
release:
	@echo "Releases are now automated via GitHub Actions"
	@echo "To create a release:"
	@echo "  1. Ensure you are on the main branch with a clean working tree"
	@echo "  2. Create and push a tag: git tag v1.2.3 && git push origin v1.2.3"
	@echo "  3. GitHub Actions will automatically build, test, and release"
	@echo ""
	@echo "The workflow will:"
	@echo "  - Run all tests"
	@echo "  - Build multi-platform binaries with GoReleaser"
	@echo "  - Create a GitHub release with artifacts"
	@echo "  - Build and push Docker images"
	@echo "  - Generate and submit CLI documentation PR"
