.PHONY: docker shell deps test pacts publish-pacts get-spec-prod get-spec-local gen-models build docs package_docker_docs

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

docker:
	docker build -t replicatedhq.replicated .

shell:
	docker run --rm -it \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated

deps:
	docker run --rm \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated glide install

test:
	go test ./cli/test

pacts:
	docker build -t replicated-cli-test -f hack/Dockerfile.testing .
	docker run --rm --name replicated-cli-tests \
		-v `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicated-cli-test \
		go test -v ./pkg/...


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
get-spec-prod:
	mkdir -p gen/spec/
	curl -o gen/spec/v1.json https://api.replicated.com/vendor/v1/spec/vendor-api.json
	curl -o gen/spec/v2.json https://api.replicated.com/vendor/v2/spec/swagger.json; # TODO this is still wrong, need to find where this is hosted

# generate the swagger specs from the local replicatedcom/vendor-api repo
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
gen-models:
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i /local/gen/spec/v1.json \
		-l go \
		-o /local/gen/go/v1; \
	# TODO this will fail, see note above in get-spec-prod
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i /local/gen/spec/v2.json \
		-l go \
		-o /local/gen/go/v2;

build:
	go build \
    		${LDFLAGS} \
    		-o replicated \
    		cli/main.go
	mv replicated ${GOPATH}/bin

docs:
	go run docs/generate.go --path ./gen/docs

package_docker_docs: docs
	docker login -p $(QUAY_PASS) -u $(QUAY_USER) quay.io
	docker build -t quay.io/replicatedcom/vendor-cli-docs:release -f docs/Dockerfile .
	docker push quay.io/replicatedcom/vendor-cli-docs:release
	docker rmi -f quay.io/replicatedcom/vendor-cli-docs:release
