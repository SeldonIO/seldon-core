FROM openjdk:8u141-jdk

ENV PYTHON_VERSION "3.6.0"

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

# ZLIB1G required for installation python
RUN apt-get upgrade && \
        apt-get update && \
        apt-get install zlib1g -y && \
        apt-get install zlib1g-dev -y && \
        apt-get install libssl-dev -y

RUN curl -SLO https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
    && tar xvf Python-${PYTHON_VERSION}.tgz \
    && cd Python-${PYTHON_VERSION} \
    && ./configure --prefix=/usr/local --with-openssl \
    && make \
    && make altinstall \
    && cd / \
    && rm -rf Python-${PYTHON_VERSION}*

# Ensure python3 and pip3 are default
RUN curl -sS https://bootstrap.pypa.io/get-pip.py >>setup.py && \
        python setup.py
RUN update-alternatives --install /usr/bin/python python /usr/local/bin/python3.6 1
RUN rm -rf /usr/local/bin/pip && \
        ln -s /usr/local/bin/pip3.6 /usr/local/bin/pip

RUN pip install --upgrade pip
RUN pip install --upgrade setuptools
RUN pip install pyyaml 
RUN pip install grpcio
RUN pip install grpcio-tools
RUN apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# INSTALL GO
ENV PATH /usr/local/go/bin:$PATH
RUN wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && \
        tar -zxvf go1.13.3.linux-amd64.tar.gz && \
        mv go/ /usr/local/go
RUN curl -sL https://go.kubebuilder.io/dl/2.0.1/linux/amd64 | tar -xz -C /tmp/ && \
        mv /tmp/kubebuilder_2.0.1_linux_amd64 /usr/local/kubebuilder/
ENV PATH /usr/local/kubebuilder/bin:$PATH

# DOWNLOAD HELM
RUN curl -L https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

# DOWNLOAD KUSTOMIZE
RUN wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.2.3/kustomize_kustomize.v3.2.3_linux_amd64 && \
        chmod +x kustomize_kustomize.v3.2.3_linux_amd64 && \
        mv kustomize_kustomize.v3.2.3_linux_amd64 /usr/local/bin/kustomize


WORKDIR /work

# Define default command.
CMD ["bash"]

