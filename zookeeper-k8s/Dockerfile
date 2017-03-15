FROM ubuntu:14.04

MAINTAINER dev@seldon.io

ENV ZOOKEEPER_VERSION 3.4.8

EXPOSE 2181 2888 3888

RUN apt-get update && apt-get -y upgrade && \
	apt-get -y install wget openjdk-7-jre-headless && \
	wget -q -O - http://apache.mirrors.pair.com/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/zookeeper-${ZOOKEEPER_VERSION}.tar.gz | tar -xzf - -C /opt && \
	mv /opt/zookeeper-${ZOOKEEPER_VERSION} /opt/zookeeper && \
	mkdir -p /opt/zookeeper/data && \
	mkdir -p /opt/zookeeper/log

WORKDIR /opt/zookeeper

VOLUME ["/opt/zookeeper/conf", "/opt/zookeeper/data", "/opt/zookeeper/log"]

COPY config-and-run.sh ./bin/

CMD ["/opt/zookeeper/bin/config-and-run.sh"]

