FROM golang:1.12

RUN curl https://glide.sh/get | sh

ENV PROJECTPATH=/go/src/github.com/replicatedhq/replicated

RUN go get golang.org/x/tools/cmd/goimports

RUN go get github.com/spf13/cobra/cobra

RUN go get github.com/go-swagger/go-swagger/cmd/swagger

WORKDIR $PROJECTPATH

CMD ["/bin/bash"]
