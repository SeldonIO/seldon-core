FROM continuumio/miniconda3:4.12.0

RUN \
    apt-get update  --allow-releaseinfo-change -y && \
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


# Install latest git
RUN apt-get upgrade -y && \
    apt-get update -y && \
    apt-get install -y make libssl-dev libghc-zlib-dev\
        libcurl4-gnutls-dev libexpat1-dev gettext unzip && \
    wget https://github.com/git/git/archive/v2.30.0.zip -O git.zip && \
    unzip git.zip && \
    apt-get remove git -y && \
    make prefix=/usr/local -C git-2.30.0/ all install


# This is to install desired version of Python without updating conda version
ENV PYTHON_VERSION "3.7.10"
ENV CONDA_VERSION "4.7.12"
RUN conda install --yes -c conda-forge python=$PYTHON_VERSION conda=$CONDA_VERSION

# Install python dependencies
RUN pip install --upgrade pip setuptools wheel && \
    pip install grpcio grpcio-tools flatbuffers twine mypy-protobuf

# Installs required
RUN apt-get install -y curl

# Install Poetry
ENV POETRY_HOME /var/poetry
RUN curl -sSL https://install.python-poetry.org | python3 - --version 1.1.15

ENV PATH "$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE false

WORKDIR /work

# Define default command.
CMD ["bash"]
