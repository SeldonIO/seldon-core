FROM openjdk:8

# Install maven
RUN apt-get update
RUN apt-get install -y maven
RUN update-alternatives --set java /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java

WORKDIR /work

ADD .m2 /work/.m2
ADD version.txt /work/version.txt

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
CMD ["/usr/lib/jvm/java-8-openjdk-amd64/bin/java", "-jar", "/work/seldon-cluster-manager.jar"]

