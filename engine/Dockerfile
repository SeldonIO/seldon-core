FROM openjdk:8u171-jre-alpine3.7

ARG APP_VERSION=UNKOWN_VERSION

#RUN apt-get update && apt-get install -y curl libblas-dev

COPY /target/seldon-engine-${APP_VERSION}.jar app.jar
COPY /target/generated-resources /licenses/

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar" ]

