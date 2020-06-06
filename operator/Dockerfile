# Build the manager binary
FROM golang:1.13.2 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY controllers/ controllers/
COPY utils/ utils/
COPY constants/ constants/
COPY client/ client/

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/manager .
COPY licenses/license.txt licenses/license.txt 
COPY generated/admissionregistration.k8s.io_v1beta1_mutatingwebhookconfiguration_seldon-mutating-webhook-configuration.yaml /tmp/operator-resources/mutate.yaml
COPY generated/admissionregistration.k8s.io_v1beta1_validatingwebhookconfiguration_seldon-validating-webhook-configuration.yaml /tmp/operator-resources/validate.yaml
COPY generated/~g_v1_service_seldon-webhook-service.yaml /tmp/operator-resources/service.yaml
COPY generated/~g_v1_configmap_seldon-config.yaml /tmp/operator-resources/configmap.yaml
COPY generated/apiextensions.k8s.io_v1beta1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml /tmp/operator-resources/crd.yaml
ENTRYPOINT ["/manager"]
