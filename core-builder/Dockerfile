FROM maven:3-jdk-8

RUN \
    apt-get update -y && \
    apt-get install -y vim && \
    apt-get install -y build-essential && \
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /work

# Define default command.
CMD ["bash"]

