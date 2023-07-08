FROM alpine:3.17

ARG RELEASE=latest

WORKDIR /app

RUN apk add curl tar gunzip && \
  curl https://github.com/thedataflows/opnsense-cli/releases/download/${RELEASE}/opnsense-cli_${RELEASE}_linux_amd64.tar.gz | tar -xzvf -

USER nonroot:nonroot

ENTRYPOINT ["/app/opnsense-cli", "sample-command"]
