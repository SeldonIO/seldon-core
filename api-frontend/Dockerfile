FROM openjdk:8u131-jre-alpine

ARG APP_VERSION=UNKOWN_VERSION

ADD /target/seldon-apife-${APP_VERSION}.jar app.jar

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar $SPRING_OPTS" ]

