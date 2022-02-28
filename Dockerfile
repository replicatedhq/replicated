FROM golang:1.17

ENV PROJECTPATH=/go/src/github.com/replicatedhq/replicated

RUN go get github.com/go-swagger/go-swagger/cmd/swagger

WORKDIR $PROJECTPATH

ENV GO111MODULE=on

CMD ["/bin/bash"]
