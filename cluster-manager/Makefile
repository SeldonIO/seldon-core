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
	@rm -fr src/main/proto/*

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify

update_proto:
	cp -v ../proto/seldon_deployment.proto src/main/proto/
	cp -vr ../proto/k8s/k8s.io src/main/proto

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

code_coverage: update_proto
	mvn clean verify
	cd target/site && aws s3 cp jacoco s3://seldon-core-docs/master/cluster-manager/jacoco --recursive --acl public-read


dependencies: update_proto
	mvn clean project-info-reports:dependencies
	cd target/site && aws s3 cp dependencies.html s3://seldon-core-docs/master/cluster-manager/dependencies/dependencies.html --acl public-read
	cd target/site && aws s3 cp css s3://seldon-core-docs/master/cluster-manager/dependencies/css --recursive --acl public-read
	cd target/site && aws s3 cp images s3://seldon-core-docs/master/cluster-manager/dependencies/images --recursive --acl public-read
