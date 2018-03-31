IMAGE_NAME=seldon-java-wrapper
VERSION_FILE=target/version.txt

build_jar: update_proto
	mvn clean package -B

write_version:
	ls target/seldon-core-wrapper-*.jar | sed -n 's/target\/seldon-core-wrapper-\(.*\).jar$$/\1/p' > $(VERSION_FILE) && cat $(VERSION_FILE)

build_image: build_jar write_version
	docker build --build-arg APP_VERSION=$$(cat $(VERSION_FILE)) -t seldonio/$(IMAGE_NAME):latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

push_to_registry:
	docker push seldonio/$(IMAGE_NAME):$$(cat $(VERSION_FILE))

update_proto:
	@cp -v ../../../../proto/prediction.proto src/main/proto/

clean:
	rm -rf src/main/proto/*
	mvn clean
