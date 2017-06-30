FROM golang:1.8

RUN curl https://glide.sh/get | sh

ENV PROJECTPATH=/go/src/github.com/replicatedhq/replicated

RUN go get golang.org/x/tools/cmd/goimports

WORKDIR $PROJECTPATH

CMD ["/bin/bash"]
