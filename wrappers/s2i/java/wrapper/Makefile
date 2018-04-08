IMAGE_NAME=seldon-java-wrapper
VERSION_FILE=target/version.txt

build_jar: update_proto
	mvn clean package -B

update_proto:
	@cp -v ../../../../proto/prediction.proto src/main/proto/

clean:
	rm -rf src/main/proto/*
	mvn clean
