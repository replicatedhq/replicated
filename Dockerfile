FROM golang:1.8

RUN curl https://glide.sh/get | sh

ENV PROJECTPATH=/go/src/github.com/replicatedhq/replicated

RUN go get golang.org/x/tools/cmd/goimports

RUN go get github.com/spf13/cobra/cobra

RUN go get github.com/go-swagger/go-swagger/cmd/swagger

RUN curl --location -o goreleaser.deb https://github.com/goreleaser/goreleaser/releases/download/v0.24.0/goreleaser_Linux_x86_64.deb
RUN dpkg -i goreleaser.deb

WORKDIR $PROJECTPATH

CMD ["/bin/bash"]
