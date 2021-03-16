FROM continuumio/miniconda3:4.7.12

RUN \
    apt-get update -y && \
    apt-get install -y vim && \
    apt-get install -y git && \
    apt-get install -y build-essential && \
    apt-get install -y cmake && \
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install python dependencies
RUN pip install --upgrade setuptools pip wheel
RUN pip install grpcio grpcio-tools twine mypy-protobuf

# Install latest git
RUN apt-get upgrade -y && \
    apt-get update -y && \
    apt-get install -y \
        make curl gettext unzip \
        libssl-dev libghc-zlib-dev \
        libcurl4-gnutls-dev libexpat1-dev && \
    wget https://github.com/git/git/archive/v2.30.0.zip -O git.zip && \
    unzip git.zip && \
    apt-get remove git -y && \
    make prefix=/usr/local -C git-2.30.0/ all install

WORKDIR /work

# Define default command.
CMD ["bash"]

