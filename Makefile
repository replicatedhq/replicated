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
	go test ./cli/test

# fetch the swagger specs from the production Vendor API
get-spec-prod:
	mkdir -p gen/spec/
	for PKG in ${API_PKGS}; do \
		curl -o gen/spec/$$PKG.json \
			https://api.replicated.com/vendor/v1/spec/$$PKG.json; \
	done

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
			done'

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

build:
	go build -o replicated cli/main.go
	mv replicated ${GOPATH}/bin

# release the latest tag
release:
	docker run --rm -it \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		--env GITHUB_TOKEN=${GITHUB_TOKEN} \
		replicatedhq.replicated goreleaser
