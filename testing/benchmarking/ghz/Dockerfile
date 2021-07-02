# Build the manager binary
FROM golang:1.13 as builder

ARG GHZ_VERSION
RUN git clone -b ${GHZ_VERSION} https://github.com/bojand/ghz && \
    	cd ghz && make build

RUN git clone https://github.com/SeldonIO/seldon-core

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu

WORKDIR /

RUN apt-get update && \
    	    apt-get install -y wget && \
	    wget https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && \
	    mv jq-linux64 jq && chmod a+x jq && mv jq /usr/local/bin && \
	    rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/ghz/dist/ghz /usr/local/bin
COPY --from=builder /go/seldon-core/executor/proto /proto
COPY --from=builder /go/seldon-core/executor/api/grpc/kfserving/inference/*.proto /proto

CMD ["ghz","--help"]
