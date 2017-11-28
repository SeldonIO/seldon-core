FROM openjdk:8u141-jdk

# Install maven
RUN apt-get update
RUN apt-get install -y maven
RUN update-alternatives --set java /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java

WORKDIR /work


ADD .m2 /work/.m2
ADD version.txt /work/version.txt

RUN git clone -b updated_protos https://github.com/cliveseldon/java.git java_client && \
    cd java_client && \
    mvn -Dmaven.repo.local=../.m2 install 

# Prepare by downloading dependencies
ADD pom.xml /work/pom.xml
RUN ["mvn", "-Dmaven.repo.local=./.m2", "dependency:resolve"]
RUN ["mvn", "-Dmaven.repo.local=./.m2", "verify"]

# Adding source, compile and package
ADD src /work/src
RUN ["mvn", "-Dmaven.repo.local=./.m2", "clean", "package"]

RUN \
    mv /work/target/seldon-cluster-manager.jar /work/ && \
    rm -rf /work/target && \
    rm -rf /work/src && \
    rm -rf /work/.m2 && \
    rm -rf /work/pom.xml

EXPOSE 8080
ENTRYPOINT [ "sh", "-c", "java $JAVA_OPTS -jar /work/seldon-cluster-manager.jar $SPRING_OPTS" ]

