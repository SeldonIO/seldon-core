DOCKER_IMAGE_NAME=seldonio/cluster-manager
PROJECT_VERSION:=$(shell cat version.txt)
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
	@cat cluster-manager.json.in | sed \
		-e  "s|%CLUSTER_MANAGER_IMAGE_VERSION%|$(CLUSTER_MANAGER_IMAGE_VERSION)|" > $@ && echo "created $@"

start_cluster_manager: cluster-manager.json
	kubectl create -f cluster-manager.json
stop_cluster_manager: cluster-manager.json
	kubectl delete -f cluster-manager.json


update_version_txt:
	PROJECT_VERSION=$$(MAVEN_OPTS="-Dorg.slf4j.simpleLogger.defaultLogLevel=WARN -Dorg.slf4j.simpleLogger.log.org.apache.maven.plugins.help=INFO" mvn help:evaluate -o -Dexpression=project.version | tail -1) && echo "$${PROJECT_VERSION}" > version.txt && cat version.txt

update_proto:
	cp -v ../proto/seldonengine.proto src/main/proto/

