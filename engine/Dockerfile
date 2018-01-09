FROM deeplearning4j/dl4j-devkit

ARG APP_VERSION=UNKOWN_VERSION

ADD /target/seldon-engine-${APP_VERSION}.jar app.jar

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom -jar app.jar" ]

