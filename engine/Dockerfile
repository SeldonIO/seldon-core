FROM openjdk:8u131-jre-alpine

ARG APP_VERSION=UNKOWN_VERSION

RUN apk update && apk add curl

ADD /target/seldon-engine-${APP_VERSION}.jar app.jar

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom -jar app.jar" ]

