SHELL=/bin/bash

IMAGE_NAME=cluster-manager
VERSION_FILE=target/version.txt

build_jar: update_proto 
	mvn clean verify -Dlicense.useMissingFile -B

write_version:
	ls target/seldon-cluster-manager-*.jar | sed -n 's/target\/seldon-cluster-manager-\(.*\).jar$$/\1/p' > $(VERSION_FILE) && cat $(VERSION_FILE)

build_image: build_jar write_version
	docker build --build-arg APP_VERSION=$$(cat $(VERSION_FILE)) -t seldonio/$(IMAGE_NAME):latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

push_to_registry:
	docker push seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

clean:
	@mvn clean
	@rm -fv cluster-manager.json

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify

update_proto:
	cp -v ../proto/seldon_deployment.proto src/main/proto/
	cp -vr ../proto/k8s/k8s.io src/main/proto
	cp -v ../proto/k8s/v1.proto src/main/proto

cluster-manager-client-secret.txt:
	@openssl rand -base64 12 > $@ && echo "created: $@"


port_forward_cluster_manager:
	POD_NAME=$$(kubectl --namespace default get pod -l app=seldon-cluster-manager-server -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 8080:8080
check:
	curl -s -i localhost:8080/ping

port_forward_cluster_manager_debug:
	POD_NAME=$$(kubectl --namespace default get pod -l app=seldon-cluster-manager-server -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 8000:8000

get_published_versions:
	curl -L -s 'https://registry.hub.docker.com/v2/repositories/seldonio/cluster-manager/tags' | jq '.results[]|.name'


show_logs_cluster_manager:
	POD_NAME=$$(kubectl --namespace default get pod -l app=seldon-cluster-manager-server -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl logs $${POD_NAME} -f
