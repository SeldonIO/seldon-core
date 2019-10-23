FROM openjdk:8u141-jdk

# install docker
RUN \
    apt-get update && \
    apt-get install -y \
         apt-transport-https \
         ca-certificates \
         curl \
         gnupg2 \
         software-properties-common && \
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add - && \
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
       $(lsb_release -cs) \
       stable" && \
    apt-get update && \
    apt-get install -y docker-ce

# Installing source to image (s2i)
RUN wget https://github.com/openshift/source-to-image/releases/download/v1.1.14/source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    tar -zxvf source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    rm source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    mv ./s2i /bin && \
    mv ./sti /bin

# Installing KIND
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-linux-amd64 && \
    chmod +x kind-linux-amd64 && \
    mv ./kind-linux-amd64 /bin/kind

# Installing Kubectl
RUN wget https://storage.googleapis.com/kubernetes-release/release/v1.16.2/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /bin

RUN \
    apt-get update -y && \
    apt-get install -y vim && \
    apt-get install -y build-essential && \
    apt-get install -y maven && \
    update-alternatives --set java /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java && \
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# dependencies for release script
RUN \
    # install env for python3
    apt-get update -y && \
    apt-get install -y python3-pip && \
    pip3 install pyyaml && \
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Ensure pip can be reached
RUN ln -s /usr/bin/python3 python \
  && pip3 install --upgrade pip

WORKDIR /work

# Define default command.
CMD ["bash"]

