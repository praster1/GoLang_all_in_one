FROM golang:1.27.3-alpine3.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY examples/28-build-deploy/cli ./examples/28-build-deploy/cli
RUN CGO_ENABLED=0 go test ./examples/28-build-deploy/cli
RUN CGO_ENABLED=0 go build -trimpath -o /out/hello ./examples/28-build-deploy/cli

FROM scratch
COPY --from=build /out/hello /hello
USER 65532:65532
ENTRYPOINT ["/hello"]
CMD ["-name", "Go"]
