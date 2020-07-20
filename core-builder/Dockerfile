FROM openjdk:13.0.1-jdk-buster

ENV PYTHON_VERSION "3.7.4"

# install docker
RUN \
    apt-get update && \
    apt-get install -y \
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
    apt-get install -y docker-ce

# Installing source to image (s2i)
RUN wget https://github.com/openshift/source-to-image/releases/download/v1.1.14/source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    tar -zxvf source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    rm source-to-image-v1.1.14-874754de-linux-386.tar.gz && \
    mv ./s2i /bin && \
    mv ./sti /bin

# Installing KIND
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-linux-amd64 && \
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
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install maven
ENV MAVEN_VERSION=3.6.3
ENV USER_HOME_DIR="/root"
ENV SHA=c35a1803a6e70a126e80b2b3ae33eed961f83ed74d18fcd16909b2d44d7dada3203f1ffe726c17ef8dcca2dcaa9fca676987befeadc9b9f759967a8cb77181c0
ENV BASE_URL=https://apache.osuosl.org/maven/maven-3/${MAVEN_VERSION}/binaries
RUN mkdir -p /usr/share/maven /usr/share/maven/ref \
  && curl -fsSL -o /tmp/apache-maven.tar.gz ${BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz \
  && echo "${SHA}  /tmp/apache-maven.tar.gz" | sha512sum -c - \
  && tar -xzf /tmp/apache-maven.tar.gz -C /usr/share/maven --strip-components=1 \
  && rm -f /tmp/apache-maven.tar.gz \
  && ln -s /usr/share/maven/bin/mvn /usr/bin/mvn
ENV MAVEN_HOME=/usr/share/maven
ENV MAVEN_CONFIG="$USER_HOME_DIR/.m2"

# compression and ssl libs required for installation python and dependencies (pandas)
RUN apt-get upgrade && \
        apt-get update && \
        apt-get dist-upgrade -y && \
        apt-get install libncursesw5-dev libgdbm-dev libc6-dev -y && \
        apt-get install zlib1g -y && \
        apt-get install zlib1g-dev libsqlite3-dev tk-dev -y && \
        apt-get install openssl libssl-dev -y && \
	apt-get install libbz2-dev -y && \
	apt-get install liblzma-dev -y && \
    apt-get install libffi-dev -y

ENV CONDA_VERSION=4.7.12

RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-${CONDA_VERSION}-Linux-x86_64.sh -O ~/miniconda.sh && \
    /bin/bash ~/miniconda.sh -b -p /opt/conda && \
    rm ~/miniconda.sh && \
    /opt/conda/bin/conda clean -tipsy && \
    ln -s /opt/conda/etc/profile.d/conda.sh /etc/profile.d/conda.sh && \
    echo ". /opt/conda/etc/profile.d/conda.sh" >> ~/.bashrc && \
    echo "conda activate base" >> ~/.bashrc && \
    find /opt/conda/ -follow -type f -name '*.a' -delete && \
    find /opt/conda/ -follow -type f -name '*.js.map' -delete && \
    /opt/conda/bin/conda clean -afy

# INSTALL GO
ENV GO_VERSION=1.14.6
ENV PATH /usr/local/go/bin:$PATH
RUN wget "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" && \
        tar -zxvf "go$GO_VERSION.linux-amd64.tar.gz" && \
        mv go/ /usr/local/go
RUN curl -sL https://go.kubebuilder.io/dl/2.3.0/linux/amd64 | tar -xz -C /tmp/ && \
        mv /tmp/kubebuilder_2.3.0_linux_amd64 /usr/local/kubebuilder/
ENV PATH /usr/local/kubebuilder/bin:$PATH

# DOWNLOAD HELM
RUN curl -L https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash -s -- --version v3.2.0

# DOWNLOAD KUSTOMIZE
RUN wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.2.3/kustomize_kustomize.v3.2.3_linux_amd64 && \
        chmod +x kustomize_kustomize.v3.2.3_linux_amd64 && \
        mv kustomize_kustomize.v3.2.3_linux_amd64 /usr/local/bin/kustomize

# iptables fix
RUN \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy

# Make RUN commands use the CONDA environment - see https://pythonspeed.com/articles/activate-conda-dockerfile/
SHELL ["/opt/conda/bin/conda", "run", "-n", "base", "/bin/bash", "-c"]

# install jupyter
RUN pip install jupyter

# Ginkgo
RUN go get github.com/onsi/ginkgo/ginkgo

# Install grpcurl
RUN wget https://github.com/fullstorydev/grpcurl/releases/download/v1.6.1/grpcurl_1.6.1_linux_x86_64.tar.gz && \
        tar -zxvf grpcurl_1.6.1_linux_x86_64.tar.gz && \
        chmod +x grpcurl && \
        mv grpcurl /usr/local/bin/grpcurl

ENV PATH="${PATH}:/root/go/bin"

# Notebooks python requirements
COPY dev_requirements.txt /tmp/dev_requirements.txt
RUN pip install -r /tmp/dev_requirements.txt
# General python requirements
COPY requirements.txt /tmp/requirements.txt
COPY requirements-dev.txt /tmp/requirements-dev.txt
RUN cd /tmp && pip install -r requirements-dev.txt

# Install Jenkins X CLI
RUN curl -L "https://github.com/jenkins-x/jx/releases/download/v2.1.95/jx-darwin-amd64.tar.gz" | tar xzv "jx" && \
        chmod +x ./jx && \
        mv ./jx /usr/bin/jx

WORKDIR /work

# Define default command.
CMD ["bash"]
