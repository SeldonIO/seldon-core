# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY proto/ proto/
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY predictor/ predictor/
COPY logger/ logger/
COPY k8s/ k8s/

# Build
RUN go build -a -o executor main.go

# Copy OpenAPI folder and change the permissions
COPY api/rest/openapi/ /openapi/
RUN chmod -R 660 /openapi/

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/executor .
COPY licenses/license.txt licenses/license.txt
# Copy openapi spec and swagger UI files
COPY --from=builder /openapi/ /openapi/

ENTRYPOINT ["/executor"]
