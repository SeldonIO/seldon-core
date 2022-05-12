FROM golang:alpine as builder

RUN apk add --upgrade make cmake

WORKDIR /build
ADD . .

RUN rm -r apis && mv apis-TEMP apis
RUN make build

################################################################################

FROM golang:alpine
COPY --from=builder /build/bin/hodometer /bin/hodometer
CMD ["/bin/hodometer"]
