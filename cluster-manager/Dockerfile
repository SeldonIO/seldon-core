FROM openjdk:8

# Install maven
RUN apt-get update
RUN apt-get install -y maven
RUN update-alternatives --set java /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java

WORKDIR /code

Add .m2 /code/.m2

# Prepare by downloading dependencies
ADD pom.xml /code/pom.xml
RUN ["mvn", "-Dmaven.repo.local=./.m2", "dependency:resolve"]
RUN ["mvn", "-Dmaven.repo.local=./.m2", "verify"]

# Adding source, compile and package
ADD src /code/src
RUN ["mvn", "-Dmaven.repo.local=./.m2", "clean", "package"]

EXPOSE 8080
CMD ["/usr/lib/jvm/java-8-openjdk-amd64/bin/java", "-jar", "target/seldon-cluster-manager.jar"]

