# Build the manager binary in a build stage, then copy into a minimal runtime image.

# Build stage
FROM golang:1.22 AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

WORKDIR /workspace

# Copy vendor directory – avoids any network access in the build sandbox.
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/

COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=vendor \
    -ldflags "-s -w -X github.com/isning/redroid-operator/cmd.Version=${VERSION}" \
    -a -o manager ./cmd/

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
