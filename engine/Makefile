IMAGE_NAME=engine
VERSION_FILE=target/version.txt

build_jar: update_proto
	mvn clean package -B

write_version:
	ls target/seldon-engine-*.jar | sed -n 's/target\/seldon-engine-\(.*\).jar$$/\1/p' > $(VERSION_FILE) && cat $(VERSION_FILE)

build_image: build_jar write_version
	docker build --build-arg APP_VERSION=$$(cat $(VERSION_FILE)) -t seldonio/$(IMAGE_NAME):latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

push_to_registry:
	docker push seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE)

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
	kubectl delete --ignore-not-found=true -f engine.json


update_proto:
	@cp -v ../proto/seldon_deployment.proto src/main/proto/
	@cp -v ../proto/prediction.proto src/main/proto/
	cp -vr ../proto/k8s/k8s.io src/main/proto
	cp -v ../proto/k8s/v1.proto src/main/proto
