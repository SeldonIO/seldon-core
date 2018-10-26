

# Creates a base swagger definition
seldon.grpcapi.swagger.json:
	protoc -I/usr/local/include -I.   -I${GOPATH}/src   -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway  --swagger_out=logtostderr=true:. prediction.grpcapi.proto

engine.oa3.json apife.oa3.josn component.oa3.json:
	python create_openapis.py 

clean:
	rm -f engine.oa3.json
	rm -f apife.oa3.josn
	rm -f component.oa3.json

