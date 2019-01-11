ARG SELDON_ENGINE_IMAGE=seldonio/engine:0.5-SNAPSHOT
FROM $SELDON_ENGINE_IMAGE

RUN wget https://www.yourkit.com/download/docker/YourKit-JavaProfiler-2018.04-docker.zip -P /tmp/ && \
  unzip /tmp/YourKit-JavaProfiler-2018.04-docker.zip -d /usr/local && \
    rm /tmp/YourKit-JavaProfiler-2018.04-docker.zip

RUN apk add --no-cache libc6-compat


ENTRYPOINT [ "sh", "-c", "java -agentpath:/usr/local/YourKit-JavaProfiler-2018.04/bin/linux-x86-64/libyjpagent.so=listen=all -Djava.security.egd=file:/dev/./urandom $JAVA_OPTS -jar app.jar" ]