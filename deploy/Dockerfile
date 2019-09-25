FROM alpine:latest
RUN apk add --no-cache ca-certificates curl git nodejs npm && update-ca-certificates
RUN npm install -g replicated-lint
COPY replicated /replicated
ENV IN_CONTAINER 1
LABEL "com.replicated.vendor_cli"="true"
WORKDIR /out
ENTRYPOINT [ "/replicated" ]

