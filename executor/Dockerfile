# Build the manager binary
FROM golang:1.13 as builder
ARG base=.

WORKDIR /workspace
# Copy the Go Modules manifests
COPY ${base}/go.mod go.mod
COPY ${base}/go.sum go.sum
COPY ${base}/proto/ proto/
COPY operator/ ../operator/
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY ${base}/main.go main.go
COPY ${base}/api/ api/
COPY ${base}/predictor/ predictor/
COPY ${base}/logger/ logger/
COPY ${base}/k8s/ k8s/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o executor main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
ARG base=.
WORKDIR /
COPY --from=builder /workspace/executor .
COPY ${base}/licenses/license.txt .
ENTRYPOINT ["/executor"]
