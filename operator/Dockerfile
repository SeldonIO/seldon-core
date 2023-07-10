# Build the manager binary
FROM golang:1.19.10 as builder

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

FROM registry.access.redhat.com/ubi9/ubi-micro:9.2-9
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
