.PHONY: docs

API_PKGS=apps channels releases

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
	cd cli/test && go test

# fetch the swagger specs from the production Vendor API
get-spec-prod:
	mkdir -p gen/spec/
	for PKG in ${API_PKGS}; do \
		curl -o gen/spec/$$PKG.json \
			https://api.replicated.com/vendor/v1/spec/$$PKG.json; \
	done
	curl -o gen/spec/v2.json https://api.replicated.com/vendor/v2/spec/swagger.json;

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
	for PKG in ${API_PKGS}; do \
		docker run --rm \
			--volume `pwd`:/local \
			swaggerapi/swagger-codegen-cli generate \
			-Dmodels -DmodelsDocs=false \
			-i /local/gen/spec/$$PKG.json \
			-l go \
			-o /local/gen/go/$$PKG; \
	done
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i /local/gen/spec/v2.json \
		-l go \
		-o /local/gen/go/v2;

build:
	go build -o replicated cli/main.go
	mv replicated ${GOPATH}/bin

docs:
	go run docs/generate.go --path ./gen/docs

package_docker_docs: docs
	docker login -p $(QUAY_PASS) -u $(QUAY_USER) quay.io
	docker build -t quay.io/replicatedcom/vendor-cli-docs:release -f docs/Dockerfile .
	docker push quay.io/replicatedcom/vendor-cli-docs:release
	docker rmi -f quay.io/replicatedcom/vendor-cli-docs:release
