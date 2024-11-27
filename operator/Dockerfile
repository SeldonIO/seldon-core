# Build the manager binary
FROM golang:1.22 as builder

WORKDIR /workspace
COPY . .

# Build the binary
RUN apt-get install make
RUN make -C operator  build

FROM registry.access.redhat.com/ubi9/ubi-micro:9.5
WORKDIR /
COPY --from=builder /workspace/operator/bin/manager .
USER 65532:65532

COPY operator/licenses/ /licenses/

ENTRYPOINT ["/manager"]
