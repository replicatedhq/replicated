API_PKGS=apps channels releases

BUILDTAGS = containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp

export GO111MODULE=on

export CGO_ENABLED=0

.PHONY: test-unit
test-unit:
	go test -v `go list ./... | grep -v /pact` -tags "$(BUILDTAGS)"

.PHONY: test-pact
test-pact:
	go test -v ./pact/... -tags "$(BUILDTAGS)"

.PNONY: test-e2e
test-e2e:
	# integration and e2e
	docker build -t replicated-cli-test -f hack/Dockerfile.testing .
	docker run --rm --name replicated-cli-tests \
		-v `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicated-cli-test

.PHONY: test
test: test-unit test-pact test-e2e

.PHONY: publish-pact
publish-pact:
	pact-broker publish ./pacts \
		--auto-detect-version-properties \
		--consumer-app-version ${PACT_VERSION} \
		--verbose

.PHONY: can-i-deploy
can-i-deploy:
	pact-broker can-i-deploy \
		--pacticipant replicated-cli \
		--version ${PACT_VERSION} \
		--to-environment production \
		--verbose

.PHONY: record-release
record-release:
	pact-broker record-release \
		--pacticipant replicated-cli \
		--version ${PACT_VERSION} \
		--environment production \
		--verbose

.PHONY: record-support-ended
record-support-ended:
	pact-broker record-support-ended \
		--pacticipant replicated-cli \
		--version ${PACT_VERSION} \
		--environment production \
		--verbose

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

.PHONY: docs
docs:
	go run ./docs/

.PHONE: release
release:
	dagger call release \
		--one-password-service-account-production env:OP_SERVICE_ACCOUNT_PRODUCTION \
		--version $(version) \
		--github-token env:GITHUB_TOKEN \
		--progress plain
