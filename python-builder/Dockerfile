FROM continuumio/miniconda3:4.7.12

RUN \
    apt-get update -y && \
    apt-get install -y vim && \
    apt-get install -y git && \
    apt-get install -y build-essential && \
    apt-get install -y cmake && \
    apt-get remove -y --auto-remove && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


# Install flatc
RUN INSTALL_DIR=/tmp/flatc-install && \
    mkdir $INSTALL_DIR && \
    cd $INSTALL_DIR && \
    git clone https://github.com/google/flatbuffers.git && \
    cd flatbuffers && \
    cmake -G "Unix Makefiles" && \
    make && \
    cp ./flatc /usr/local/bin/ && \
    rm -rf $INSTALL_DIR


# Install python dependencies
RUN pip install grpcio
RUN pip install grpcio-tools
RUN pip install flatbuffers
RUN pip install twine
RUN pip install mypy-protobuf

# Install latest git
RUN apt-get upgrade -y
RUN apt-get update -y
RUN apt-get install -y make libssl-dev libghc-zlib-dev libcurl4-gnutls-dev libexpat1-dev gettext unzip
RUN wget https://github.com/git/git/archive/v2.30.0.zip -O git.zip
RUN unzip git.zip
RUN apt-get remove git -y
RUN make prefix=/usr/local -C git-2.30.0/ all install

# Installs requride
RUN apt-get install curl

WORKDIR /work

# Define default command.
CMD ["bash"]

