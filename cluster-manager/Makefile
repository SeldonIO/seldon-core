SHELL=/bin/bash

DOCKER_IMAGE_NAME=seldonio/cluster-manager
PROJECT_VERSION=$(shell cat version.txt)
DOCKER_IMAGE_VERSION=$(PROJECT_VERSION)
CLUSTER_MANAGER_IMAGE_VERSION=$(PROJECT_VERSION)

build_docker_image: update_proto update_version_txt
	mkdir -p .m2 && docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

clean:
	@mvn clean
	@rm -fv cluster-manager.json

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify

FORCE:

cluster-manager.json: FORCE
	@ \
	if [ ! -f settings.sh ]; then \
		cp -v settings.sh.default settings.sh; \
	fi && \
	source settings.sh && cat cluster-manager.json.in | sed \
		-e  "s|%SELDON_CLUSTER_MANAGER_SERVICE_TYPE%|$${SELDON_CLUSTER_MANAGER_SERVICE_TYPE}|" \
		-e  "s|%JAVA_OPTS%|$${JAVA_OPTS}|" \
		-e  "s|%SPRING_OPTS%|$${SPRING_OPTS}|" \
		-e  "s|%CLUSTER_MANAGER_IMAGE_VERSION%|$(CLUSTER_MANAGER_IMAGE_VERSION)|" > $@ && echo "created $@"

start_cluster_manager: cluster-manager.json cluster-manager-client-secret.txt
	@CLUSTER_MANAGER_CLIENT_SECRET="$$(cat cluster-manager-client-secret.txt)" && \
			kubectl delete secret cluster-manager-client-secret  > /dev/null 2>&1 || : && \
			kubectl create secret generic cluster-manager-client-secret --from-literal=cluster-manager-client-secret.txt="$${CLUSTER_MANAGER_CLIENT_SECRET}"
	@kubectl apply -f cluster-manager.json
stop_cluster_manager: cluster-manager.json
	@kubectl delete --ignore-not-found=true -f cluster-manager.json


update_version_txt:
	PROJECT_VERSION=$$(MAVEN_OPTS="-Dorg.slf4j.simpleLogger.defaultLogLevel=WARN -Dorg.slf4j.simpleLogger.log.org.apache.maven.plugins.help=INFO" mvn help:evaluate -o -Dexpression=project.version | tail -1) && echo "$${PROJECT_VERSION}" > version.txt && cat version.txt

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
