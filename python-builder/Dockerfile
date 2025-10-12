FROM condaforge/miniforge3:4.12.0-2

ENV DEBIAN_FRONTEND=noninteractive

RUN \
    apt-get update -y && \
    apt-get install -y vim git build-essential cmake && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


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


# This is to install desired version of Python without updating conda version
ENV PYTHON_VERSION="3.7.10"
ENV CONDA_VERSION="4.7.12"
RUN conda install --yes -c conda-forge python=$PYTHON_VERSION conda=$CONDA_VERSION

# Install python dependencies
RUN pip3 install --upgrade pip setuptools wheel && \
    pip3 install grpcio grpcio-tools flatbuffers twine mypy-protobuf

# Installs required
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Install Poetry
ENV POETRY_HOME=/var/poetry
RUN curl -sSL https://install.python-poetry.org | python3 - --version 1.1.15

ENV PATH="$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE=false

WORKDIR /work

# Define default command.
CMD ["bash"]
