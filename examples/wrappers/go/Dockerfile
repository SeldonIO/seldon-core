FROM golang

RUN go get -u google.golang.org/grpc

RUN go get -u github.com/gorilla/mux

WORKDIR /go/src/github.com/seldonio/seldon-core/examples/wrappers/go

COPY . .

RUN go build -o /server

ENTRYPOINT [ "sh", "-c", "/server --server_type ${SERVER_TYPE:-grpc}" ]

