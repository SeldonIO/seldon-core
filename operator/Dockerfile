# Build the manager binary
FROM golang:1.14.9 as builder

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
RUN go build -a -o manager main.go

# Get MPL licensed dependencies
RUN wget -O hashicorp-golang-lru.tar.gz https://github.com/hashicorp/golang-lru/archive/master.tar.gz
RUN wget -O armon-consul-api.tar.gz https://github.com/armon/consul-api/archive/master.tar.gz
RUN wget -O hasicorp-hcl.tar.gz https://github.com/hashicorp/hcl/archive/hcl1.tar.gz

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base:latest
WORKDIR /
COPY --from=builder /workspace/manager .
COPY licenses/license.txt licenses/license.txt
COPY --from=builder /workspace/hashicorp-golang-lru.tar.gz licenses/mpl_source/hashicorp-golang-lru.tar.gz
COPY --from=builder /workspace/armon-consul-api.tar.gz licenses/mpl_source/armon-consul-api.tar.gz
COPY --from=builder /workspace/hasicorp-hcl.tar.gz licenses/mpl_source/hasicorp-hcl.tar.gz
COPY generated/admissionregistration.k8s.io_v1beta1_validatingwebhookconfiguration_seldon-validating-webhook-configuration.yaml /tmp/operator-resources/validate.yaml
COPY generated/v1_service_seldon-webhook-service.yaml /tmp/operator-resources/service.yaml
COPY generated/v1_configmap_seldon-config.yaml /tmp/operator-resources/configmap.yaml
COPY generated/apiextensions.k8s.io_v1beta1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml /tmp/operator-resources/crd-v1beta1.yaml
COPY generated/apiextensions.k8s.io_v1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml /tmp/operator-resources/crd-v1.yaml
ENTRYPOINT ["/manager"]
