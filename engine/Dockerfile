FROM maven:3.6.3-jdk-11 as builder

WORKDIR /workspace

COPY pom.xml pom.xml
COPY src/ src/

# Build
RUN mvn clean verify -Dlicense.useMissingFile -B

FROM openjdk:11.0.5-jre-slim

COPY --from=builder /workspace/target/seldon-engine-*.jar app.jar
COPY --from=builder /workspace/target/generated-resources /licenses/

ENTRYPOINT [ "sh", "-c", "java -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar" ]

