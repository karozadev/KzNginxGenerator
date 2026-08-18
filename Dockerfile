# This Dockerfile is consumed by GoReleaser during the release pipeline,
# which builds the `kznginx` binary for each target platform first and
# runs `docker build` with that binary copied into the build context
# (see .goreleaser.yml `dockers:`). It expects a prebuilt ./kznginx
# binary next to it, so it is not meant to be built standalone with
# `docker build .` from the repository root — use `make build` +
# `go run .` for local development instead.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY kznginx /usr/local/bin/kznginx

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/kznginx"]
CMD ["ui", "--port", "8080"]
