docker:
	docker build -t replicatedhq.replicated .

shell:
	docker run --rm -it \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated

clean:
	rm -rf gen

deps:
	docker run --rm \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated glide install

test:
	go test ./client

gen:
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i https://api.replicated.com/vendor/v1/spec/channels.json \
		-l go \
		-o /local/gen/go/channels
	docker run --rm \
		--volume `pwd`:/local \
		swaggerapi/swagger-codegen-cli generate \
		-Dmodels -DmodelsDocs=false \
		-i https://api.replicated.com/vendor/v1/spec/releases.json \
		-l go \
		-o /local/gen/go/releases
	sudo chown -R ${USER}:${USER} gen/
	# fix time.Time fields. Codegen generates empty Time struct.
	rm gen/go/releases/time.go
	sed -i 's/Time/time.Time/' gen/go/releases/app_release_info.go
	# import "time"
	docker run --rm \
		--volume `pwd`:/go/src/github.com/replicatedhq/replicated \
		replicatedhq.replicated goimports -w gen/go/releases

build: deps gen
