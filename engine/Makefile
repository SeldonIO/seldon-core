DOCKER_IMAGE_NAME=seldonio/engine
PROJECT_VERSION:=`cat version.txt`
DOCKER_IMAGE_VERSION=$(PROJECT_VERSION)
ENGINE_IMAGE_VERSION=$(PROJECT_VERSION)

build_docker_image: update_proto
	mkdir -p .m2 && docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

clean:
	@mvn clean
	@rm -fv engine.json

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify



engine.json: engine.json.in Makefile
	@cat engine.json.in | sed \
		-e  "s|%ENGINE_IMAGE_VERSION%|$(ENGINE_IMAGE_VERSION)|" > $@ && echo "created $@"

start_engine: engine.json
	kubectl create -f engine.json
stop_engine: engine.json
	kubectl delete -f engine.json


update_version_txt:
	PROJECT_VERSION=$$(MAVEN_OPTS="-Dorg.slf4j.simpleLogger.defaultLogLevel=WARN -Dorg.slf4j.simpleLogger.log.org.apache.maven.plugins.help=INFO" mvn help:evaluate -o -Dexpression=project.version | tail -1) && echo "$${PROJECT_VERSION}" > version.txt

update_proto:
	cp -v ../proto/seldonengine.proto src/main/proto/ 
