# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY executor/go.mod go.mod
COPY executor/go.sum go.sum
COPY executor/proto/ proto/
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY executor/operator/ operator/
RUN go mod download

# Copy the go source
COPY executor/main.go main.go
COPY executor/api/ api/
COPY executor/predictor/ predictor/
COPY executor/logger/ logger/
COPY executor/k8s/ k8s/

# Build
RUN go build -a -o executor main.go

# Copy OpenAPI folder and change the permissions
COPY executor/api/rest/openapi/ /openapi/
RUN chmod -R 660 /openapi/

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/executor .
COPY executor/licenses/license.txt licenses/license.txt
# Copy openapi spec and swagger UI files
COPY --from=builder /openapi/ /openapi/

ENTRYPOINT ["/executor"]
