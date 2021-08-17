API_PKGS=apps channels releases

VERSION=$(shell git describe)
ABBREV_VERSION=$(shell git describe --abbrev=0)
VERSION_PACKAGE = github.com/replicatedhq/replicated/pkg/version
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
export GO111MODULE=on

GIT_TREE = $(shell git rev-parse --is-inside-work-tree 2>/dev/null)
ifneq "$(GIT_TREE)" ""
define GIT_UPDATE_INDEX_CMD
git update-index --assume-unchanged
endef
define GIT_SHA
`git rev-parse HEAD`
endef
else
define GIT_UPDATE_INDEX_CMD
echo "Not a git repo, skipping git update-index"
endef
define GIT_SHA
""
endef
endif

define LDFLAGS
-ldflags "\
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef

.PHONY: docker
docker:
	docker build -t replicatedhq.replicated .

.PHONY: shell
shell:
	docker run --rm -it \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated

.PHONY: deps
deps:
	docker run --rm \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated glide install

.PHONY: test-env
test-env:
	@if [ -z "${REPLICATED_API_TOKEN}" ]; then echo "Missing REPLICATED_API_TOKEN"; exit 1; fi
	@if [ -z "${VENDOR_USER_PASSWORD}" ]; then echo "Missing VENDOR_USER_PASSWORD"; exit 1; fi
	@if [ -z "${VENDOR_USER_EMAIL}" ]; then echo "Missing VENDOR_USER_EMAIL"; exit 1; fi
	@if [ -z "${REPLICATED_API_ORIGIN}" ]; then echo "Missing REPLICATED_API_ORIGIN"; exit 1; fi
	@if [ -z "${REPLICATED_ID_ORIGIN}" ]; then echo "Missing REPLICATED_ID_ORIGIN"; exit 1; fi

.PHONY: test
test: test-env
	go test -v ./cli/test

.PHONY: pacts
pacts:
	docker build -t replicated-cli-test -f hack/Dockerfile.testing .
	docker run --rm --name replicated-cli-tests \
		-v `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicated-cli-test \
		go test -v ./pkg/...


.PHONY: publish-pacts
publish-pacts:
	curl \
		--silent --output /dev/null --show-error --fail \
		--user ${PACT_BROKER_USERNAME}:${PACT_BROKER_PASSWORD} \
		-X PUT \
		-H "Content-Type: application/json" \
		-d@pacts/replicated-cli-vendor-graphql-api.json \
		https://replicated-pact-broker.herokuapp.com/pacts/provider/vendor-graphql-api/consumer/replicated-cli/version/$(ABBREV_VERSION)
	curl \
		--silent --output /dev/null --show-error --fail \
		--user ${PACT_BROKER_USERNAME}:${PACT_BROKER_PASSWORD} \
		-X PUT \
		-H "Content-Type: application/json" \
		-d@pacts/replicated-cli-kots-vendor-graphql-api.json \
		https://replicated-pact-broker.herokuapp.com/pacts/provider/vendor-graphql-api/consumer/replicated-cli-kots/version/$(ABBREV_VERSION)
	curl \
		--silent --output /dev/null --show-error --fail \
		--user ${PACT_BROKER_USERNAME}:${PACT_BROKER_PASSWORD} \
		-X PUT \
		-H "Content-Type: application/json" \
		-d@pacts/replicated-cli-vendor-api.json \
		https://replicated-pact-broker.herokuapp.com/pacts/provider/vendor-api/consumer/replicated-cli/version/$(ABBREV_VERSION)

# fetch the swagger specs from the production Vendor API
.PHONY: get-spec-prod
get-spec-prod:
	mkdir -p gen/spec/
	curl -o gen/spec/v1.json https://api.replicated.com/vendor/v1/spec/vendor-api.json
	curl -o gen/spec/v2.json https://api.replicated.com/vendor/v2/spec/vendor-api-v2.json
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
		-i /local/gen/spec/v2.json \
		-l go \
		-o /local/gen/go/v2;
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
    		-o bin/replicated \
    		cli/main.go

.PHONY: docs
docs:
	go run ./docs/
