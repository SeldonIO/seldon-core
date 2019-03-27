FROM registry.redhat.io/redhat-openjdk-18/openjdk18-openshift
MAINTAINER devext@seldon.io

LABEL name="Cluster Manager" \
      vendor="Seldon" \
      version="0.2.7-SNAPSHOT" \
      summary="Cluster Manager for Seldon Core" \
      description="One of the components in Seldon Core"

ARG APP_VERSION=UNKOWN_VERSION

COPY /target/seldon-apife-${APP_VERSION}.jar app.jar
COPY /target/generated-resources /licenses/

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar $SPRING_OPTS" ]

