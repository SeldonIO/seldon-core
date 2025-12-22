FROM condaforge/miniforge3:25.3.1-0

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
ENV PYTHON_VERSION="3.12.12"
ENV CONDA_VERSION="25.3.1"
# Pin brotli to 1.2.x to remediate GHSA-2qfp-q593-8484 (CVE-2025-6176)
# Pin urllib3 to 2.6.x to remediate CVE-2025-66418, CVE-2025-66471 
RUN conda install --yes -c conda-forge python=$PYTHON_VERSION conda=$CONDA_VERSION brotli=1.2 urllib3=2.6

# Install python dependencies
RUN pip3 install --upgrade pip setuptools wheel && \
    pip3 install \
        grpcio==1.70.0 \
        grpcio-tools==1.70.0 \
        flatbuffers==24.3.25 \
        twine==6.2.0 \
        mypy-protobuf==3.5.0

# Installs required
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Install Poetry
ENV POETRY_HOME=/var/poetry
ENV POETRY_VERSION=2.1.2
RUN curl -sSL https://install.python-poetry.org | python3 - --version $POETRY_VERSION

ENV PATH="$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE=false

WORKDIR /work

# Define default command.
CMD ["bash"]
