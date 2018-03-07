FROM openjdk:8u151-jre-slim-stretch

ARG APP_VERSION=UNKOWN_VERSION

RUN apt-get update && apt-get install -y curl libblas-dev

ADD /target/seldon-engine-${APP_VERSION}.jar app.jar

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar" ]

