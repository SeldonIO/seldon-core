
go/src/github.com/seldonio:
	mkdir -p go/src/github.com/seldonio

go/src/github.com/seldonio/seldon-operator: go/src/github.com/seldonio
	cd go/src/github.com/seldonio && git clone https://github.com/SeldonIO/seldon-operator.git

.PHONY:build
build: clean go/src/github.com/seldonio/seldon-operator
	export GOPATH=${PWD}/go && cd ./go/src/github.com/seldonio/seldon-operator && make docker-build 

.PHONY:push_image_private_repo
push_image_private_repo: go/src/github.com/seldonio/seldon-operator
	cd ./go/src/github.com/seldonio/seldon-operator && make docker-push-local-private

.PHONY:clean
clean:
	rm -rf ./go
