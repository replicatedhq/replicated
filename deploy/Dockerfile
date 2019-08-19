FROM alpine:latest
RUN apk add --no-cache ca-certificates curl git && update-ca-certificates
COPY replicated /replicated
ENV IN_CONTAINER 1
LABEL "com.replicated.vendor_cli"="true"
WORKDIR /out
ENTRYPOINT [ "/replicated" ]
