FROM continuumio/miniconda3:4.5.12

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

WORKDIR /work

# Define default command.
CMD ["bash"]

