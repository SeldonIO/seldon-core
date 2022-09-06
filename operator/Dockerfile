# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY operator/go.mod go.mod
COPY operator/go.sum go.sum
COPY components /components/
COPY apis/ /apis/
COPY operator/apis apis

# Copy the go source
COPY operator/controllers/ controllers/
COPY operator/scheduler/ scheduler/
COPY operator/pkg/ pkg/
COPY operator/main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
