FROM openjdk:13.0.1-jdk-buster

ENV PYTHON_VERSION="3.8.10"

RUN printf '%s\n' \
    'deb http://archive.debian.org/debian buster main contrib non-free' \
    'deb http://archive.debian.org/debian-security buster/updates main contrib non-free' \
    > /etc/apt/sources.list \
    && printf 'Acquire::Check-Valid-Until "false";\n' \
    'Acquire::Retries "3";\n' \
    > /etc/apt/apt.conf.d/99archive \
    && apt-get update

# install docker
RUN \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    wget \
    gnupg2 \
    software-properties-common && \
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) \
    stable" && \
    apt-get update && \
    apt-get install -y --no-install-recommends docker-ce

# Installing source to image (s2i)
RUN wget https://github.com/openshift/source-to-image/releases/download/v1.1.14/source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    tar -zxvf source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    rm source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    mv ./s2i /bin && \
    mv ./sti /bin

# Installing KIND
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/v0.14.0/kind-linux-amd64 && \
    chmod +x kind-linux-amd64 && \
    mv ./kind-linux-amd64 /bin/kind

# Installing Kubectl
RUN wget https://storage.googleapis.com/kubernetes-release/release/v1.16.2/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /bin

RUN \
    apt-get update -y && \
    apt-get install -y vim jq build-essential pandoc && \
    apt-get remove -y --auto-remove && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install maven
ENV MAVEN_VERSION=3.9.3
ENV USER_HOME_DIR="/root"
ENV SHA=400fc5b6d000c158d5ee7937543faa06b6bda8408caa2444a9c947c21472fde0f0b64ac452b8cec8855d528c0335522ed5b6c8f77085811c7e29e1bedbb5daa2
RUN set -eux; \
    BASE_URL="https://archive.apache.org/dist/maven/maven-3/${MAVEN_VERSION}/binaries"; \
    mkdir -p /usr/share/maven /usr/share/maven/ref; \
    curl -fsSL -o /tmp/apache-maven.tar.gz "${BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz"; \
    echo "${SHA}  /tmp/apache-maven.tar.gz" | sha512sum -c -; \
    tar -xzf /tmp/apache-maven.tar.gz -C /usr/share/maven --strip-components=1; \
    rm -f /tmp/apache-maven.tar.gz; \
    ln -s /usr/share/maven/bin/mvn /usr/bin/mvn
ENV MAVEN_HOME=/usr/share/maven
ENV MAVEN_CONFIG="$USER_HOME_DIR/.m2"

# compression and ssl libs required for installation python and dependencies (pandas)
RUN apt-get update && \
    apt-get install -y --no-install-recommends zlib1g zlib1g-dev libssl-dev libbz2-dev liblzma-dev && \
    rm -rf /var/lib/apt/lists/*

ENV CONDA_DOWNLOAD_VERSION="py38_23.3.1-0"
ENV CONDA_VERSION="23.5.0"
RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-${CONDA_DOWNLOAD_VERSION}-Linux-x86_64.sh -O ~/miniconda.sh && \
    /bin/bash ~/miniconda.sh -b -p /opt/conda && \
    rm ~/miniconda.sh && \
    /opt/conda/bin/conda clean -tipy && \
    ln -s /opt/conda/etc/profile.d/conda.sh /etc/profile.d/conda.sh && \
    echo ". /opt/conda/etc/profile.d/conda.sh" >> ~/.bashrc && \
    echo "conda activate base" >> ~/.bashrc && \
    chgrp -R root /opt/conda && chmod -R g+rw /opt/conda
ENV PATH="/opt/conda/bin:${PATH}"
ENV PYTHON_VERSION="3.8.10"

# This is to install desired version of Python without updating conda version
RUN conda install --yes python=$PYTHON_VERSION conda=$CONDA_VERSION

# Upgrade and install some Python packages
RUN pip install --upgrade pip setuptools wheel
RUN pip install pyyaml grpcio grpcio-tools
RUN apt-get remove -y --auto-remove \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# INSTALL GO
ENV PATH=/usr/local/go/bin:$PATH
RUN wget https://dl.google.com/go/go1.24.7.linux-amd64.tar.gz && \
    tar -zxvf go1.24.7.linux-amd64.tar.gz && \
    mv go/ /usr/local/go

# Install kubebuilder (using github link)
RUN curl -sL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.0/kubebuilder_2.3.0_linux_amd64.tar.gz | tar -xz -C /tmp/ && \
    mv /tmp/kubebuilder_2.3.0_linux_amd64 /usr/local/kubebuilder/
ENV PATH=/usr/local/kubebuilder/bin:$PATH

# DOWNLOAD HELM
RUN curl -L https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash -s -- --version v3.2.0

# DOWNLOAD KUSTOMIZE
RUN wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.1.2/kustomize_v4.1.2_linux_amd64.tar.gz && \
    tar -xzf kustomize_v4.1.2_linux_amd64.tar.gz && \
    chmod +x kustomize && \
    mv kustomize /usr/local/bin/kustomize

# iptables fix
RUN \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy

# install jupyter
RUN pip install --no-cache-dir jupyter==1.1.1

# Install grpcurl
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

ENV PATH="${PATH}:/root/go/bin"

# Notebooks python requirements
COPY dev_requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt

# Ginkgo
RUN go install github.com/onsi/ginkgo/ginkgo@v1.16.5

# Helm docs
RUN go install github.com/norwoodj/helm-docs/cmd/helm-docs@f66fdbd6fe

# Argo workflows CLI
RUN wget https://github.com/argoproj/argo-workflows/releases/download/v3.0.8/argo-linux-amd64.gz && \
    gunzip argo-linux-amd64.gz && \
    mv argo-linux-amd64 argo && \
    chmod a+x argo && \
    mv argo /usr/local/bin/argo

# Installing jx
RUN wget https://github.com/jenkins-x/jx-cli/releases/download/v3.1.242/jx-cli-linux-amd64.tar.gz && \
    tar -zxvf jx-cli-linux-amd64.tar.gz && \
    chmod a+x jx && \
    mv jx /usr/local/bin/jx

RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-349.0.0-linux-x86_64.tar.gz && \
    tar xvfz google-cloud-sdk-349.0.0-linux-x86_64.tar.gz -C /tmp/ && \
    mv /tmp/google-cloud-sdk /usr/local/google-cloud-sdk
ENV PATH=/usr/local/google-cloud-sdk/bin:$PATH

WORKDIR /work

# Define default command.
CMD ["bash"]
