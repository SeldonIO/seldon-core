SHELL=/bin/bash

IMAGE_NAME=apife
VERSION_FILE=target/version.txt

write_version:
	ls target/seldon-apife-*.jar | sed -n 's/target\/seldon-apife-\(.*\).jar$$/\1/p' > $(VERSION_FILE) && cat $(VERSION_FILE)

build_jar: update_proto update_swagger
	mvn clean verify -Dlicense.useMissingFile -B

build_image: build_jar write_version
	mkdir -p .m2 && docker build --build-arg APP_VERSION=$$(cat $(VERSION_FILE)) -t seldonio/$(IMAGE_NAME):latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))


push_to_registry:
	docker push seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

clean:
	@mvn clean
	@rm -fv apife.json
	@rm -fv apife.yaml
	@rm -fv apife_istio.yaml

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify

FORCE:


start_apife: apife.json
	kubectl apply -f apife.json
stop_apife: apife.json
	kubectl delete --ignore-not-found=true -f apife.json

start_apife_istio: apife_istio.json
	kubectl apply -f apife_istio.json

stop_apife_istio: apife_istio.json
	kubectl delete --ignore-not-found=true -f apife_istio.json


update_proto:
	cp -v ../proto/seldon_deployment.proto src/main/proto/ 
	cp -v ../proto/prediction.proto src/main/proto/ 
	cp -vr ../proto/k8s/k8s.io src/main/proto
	cp -v ../proto/k8s/v1.proto src/main/proto

port_forward_api_server:
	POD_NAME=$$(kubectl --namespace default get pod -l app=seldon-apiserver-container-app -o template --template="{{(index .items 0).metadata.name}}") && \
		kubectl port-forward $${POD_NAME} 8080:8080

update_swagger:
	cp -v ../openapi/apife.oa3.json src/main/resources/static/seldon.json
