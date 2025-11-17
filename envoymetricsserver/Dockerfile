FROM golang:1.25-alpine as builder

WORKDIR /build
COPY . .

RUN go build main.go

FROM registry.access.redhat.com/ubi9/ubi-micro:9.6
COPY --from=builder /build/main /bin/metrics


ENTRYPOINT ["/bin/metrics"]
