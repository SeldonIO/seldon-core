FROM openjdk:8u171-jre-alpine3.7

ARG APP_VERSION=UNKOWN_VERSION

COPY /target/seldon-cluster-manager-${APP_VERSION}.jar app.jar
COPY /target/generated-resources /licenses/

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar $SPRING_OPTS" ]

