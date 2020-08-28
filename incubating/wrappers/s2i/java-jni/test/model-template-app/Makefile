VERSION := 0.1.0
IMAGE_NAME_BASE=jni-model-template
IMAGE=seldonio/${IMAGE_NAME_BASE}:${VERSION}

build_jar: 
	mvn clean package -B

build:
	s2i build . \
		seldonio/java-jni-server-build:0.3.0 \
		--runtime-image seldonio/java-jni-server-runtime:0.3.0 \
		${IMAGE}

clean:
	mvn clean


