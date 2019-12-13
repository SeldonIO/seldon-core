FROM openjdk:11.0.5-jre-slim

ARG APP_VERSION=UNKOWN_VERSION

RUN \
    apt-get update -y && \
    apt-get remove -y --auto-remove && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY /target/seldon-engine-${APP_VERSION}.jar app.jar
COPY /target/generated-resources /licenses/

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar" ]

