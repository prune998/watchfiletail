FROM golang:1.21-alpine as builder

LABEL vendor="Prune - prune@lecentre.net" \
      content="watchfiletail"

ARG VERSION="0.1"
ARG BUILDTIME="20230105"

ARG TARGETOS
ARG TARGETARCH

COPY . /go/src/github.com/prune998/watchfiletail
WORKDIR /go/src/github.com/prune998/watchfiletail

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH}  go build -a -v -ldflags "-X main.version=${VERSION}-${BUILDTIME}" -o watchfiletail

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source https://github.com/prune998/watchfiletail

WORKDIR /
COPY --from=builder /go/src/github.com/prune998/watchfiletail/watchfiletail .
USER 65532:65532

ENTRYPOINT ["/watchfiletail"]
