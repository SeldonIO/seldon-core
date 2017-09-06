DOCKER_IMAGE_NAME=seldonio/apife
PROJECT_VERSION:=`cat version.txt`
DOCKER_IMAGE_VERSION=$(PROJECT_VERSION)
APIFE_IMAGE_VERSION=$(PROJECT_VERSION)

build_docker_image: update_proto
	mkdir -p .m2 && docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

clean:
	@mvn clean
	@rm -fv apife.json
	@rm -fv apife.yaml
	@rm -fv apife_istio.yaml

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify



apife.json: apife.json.in Makefile
	@cat apife.json.in | sed \
		-e  "s|%APIFE_IMAGE_VERSION%|$(APIFE_IMAGE_VERSION)|" > $@ && echo "created $@"

apife_istio.json: apife_istio.json.in Makefile
	@cat apife_istio.json.in | sed \
		-e  "s|%APIFE_IMAGE_VERSION%|$(APIFE_IMAGE_VERSION)|" > $@ && echo "created $@"

start_apife: apife.json
	kubectl apply -f apife.json
stop_apife: apife.json
	kubectl delete --ignore-not-found=true -f apife.json

start_apife_istio: apife_istio.json
	kubectl apply -f apife_istio.json

stop_apife_istio: apife_istio.json
	kubectl delete --ignore-not-found=true -f apife_istio.json


update_version_txt:
	PROJECT_VERSION=$$(MAVEN_OPTS="-Dorg.slf4j.simpleLogger.defaultLogLevel=WARN -Dorg.slf4j.simpleLogger.log.org.apache.maven.plugins.help=INFO" mvn help:evaluate -o -Dexpression=project.version | tail -1) && echo "$${PROJECT_VERSION}" > version.txt

update_proto:
	cp -v ../proto/seldonengine.proto src/main/proto/ 
	cp -v ../proto/prediction.proto src/main/proto/ 

port_forward_api_server:
	POD_NAME=$$(kubectl --namespace default get pod -l app=seldon-apiserver-container-app -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 8080:8080

