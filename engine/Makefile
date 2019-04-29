IMAGE_NAME=engine
VERSION_FILE=target/version.txt

build_jar: update_proto update_swagger
	mvn clean verify -Dlicense.useMissingFile -B

write_version:
	ls target/seldon-engine-*.jar | sed -n 's/target\/seldon-engine-\(.*\).jar$$/\1/p' > $(VERSION_FILE) && cat $(VERSION_FILE)

build_image: build_jar write_version
	docker build --build-arg APP_VERSION=$$(cat $(VERSION_FILE)) -t seldonio/$(IMAGE_NAME):latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

push_to_registry:
	docker push seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

clean:
	@mvn clean
	@rm -fv engine.json
	@rm -fr src/main/proto/*

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
	cp -vr ../proto/tensorflow/tensorflow src/main/proto

update_swagger:
	cp -v ../openapi/engine.oas3.json src/main/resources/static/seldon.json

code_coverage: update_proto
	mvn clean verify
	cd target/site && aws s3 cp jacoco s3://seldon-core-docs/master/engine/jacoco --recursive --acl public-read


dependencies: update_proto
	mvn clean project-info-reports:dependencies
	cd target/site && aws s3 cp dependencies.html s3://seldon-core-docs/master/engine/dependencies/dependencies.html --acl public-read
	cd target/site && aws s3 cp css s3://seldon-core-docs/master/engine/dependencies/css --recursive --acl public-read
	cd target/site && aws s3 cp images s3://seldon-core-docs/master/engine/dependencies/images --recursive --acl public-read
