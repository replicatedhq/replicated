FROM alpine:latest

RUN apk add --no-cache ca-certificates curl git && \
    update-ca-certificates

ENV IN_CONTAINER=1

WORKDIR /out

COPY bin/replicated /replicated

LABEL com.replicated.vendor_cli=true

ENTRYPOINT ["/replicated"]
