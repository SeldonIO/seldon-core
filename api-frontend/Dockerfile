FROM openjdk:8u171-jre-alpine3.7

ARG APP_VERSION=UNKOWN_VERSION

ADD /target/seldon-apife-${APP_VERSION}.jar app.jar

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar $SPRING_OPTS" ]

