ARG VERSION
ARG BASE_IMAGE
FROM ${BASE_IMAGE}:${VERSION} as base

ARG VERSION
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.16.0-dev" \
      release="1" \
      summary="Alibi Detect Server for Seldon Core" \
      description="The Alibi Detect Server provides outlier, drift and adversarial detection services for Seldon Core"

FROM base as builder
ARG PYTHON_VERSION
ARG CONDA_VERSION

RUN microdnf update -y && \
    microdnf install -y \
      unzip \
      make \
      automake \
      gcc \
      gcc-c++

# Install Rclone Binary to be present in the image
RUN wget https://downloads.rclone.org/v1.55.1/rclone-v1.55.1-linux-amd64.zip && \
    unzip rclone-v1.55.1-linux-amd64.zip && \
    mv rclone-v1.55.1-linux-amd64/rclone /usr/bin/rclone && \
    rm -rf rclone-v1.55.1-linux-amd64.zip rclone-v1.55.1-linux-amd64

# Install Python / Conda
# Note that we need to force Conda to use the system's std-c++ library, as
# otherwise PyStan won't compile with Conda's older stdc++ library:
# https://github.com/stan-dev/pystan/issues/294#issuecomment-870711100
RUN conda install --yes -c conda-forge \
      python=${PYTHON_VERSION} \
      conda=${CONDA_VERSION} && \
    pip install --upgrade pip setuptools wheel && \
    cd /opt/conda/lib && \
      rm libstdc++.so libstdc++.so.6 && \
      ln -s /usr/lib64/libstdc++.so.6.0.29 libstdc++.so && \
      ln -s /usr/lib64/libstdc++.so.6.0.29 libstdc++.so.6

# Make home dir
RUN mkdir microservice
WORKDIR /microservice

# Install Poetry
ENV POETRY_HOME /microservice/.poetry
RUN curl -sSL https://install.python-poetry.org | python3 - --version 1.1.15

ENV PATH "$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE false

# Install the server
COPY poetry.lock pyproject.toml ./
COPY _seldon_core ./_seldon_core
RUN poetry install && rm ~/.cache/pip -rf

# Add licences
RUN mkdir /licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
RUN cp ./licenses/* /licenses

# Copy rest of the package
COPY adserver adserver
COPY README.md README.md
COPY version.txt version.txt

FROM base as final
WORKDIR /microservice

# mesa-libGL: this is to avoid "ImportError: libGL.so.1" from opencv
RUN microdnf install -y mesa-libGL
RUN microdnf update -y

COPY --from=builder /microservice /microservice
COPY --from=builder /opt/conda /opt/conda
COPY --from=builder /usr/bin/rclone /usr/bin/rclone
COPY --from=builder /licenses /licenses

# This is to have writable numba cache directory
ENV NUMBA_CACHE_DIR /tmp/numba-cache

USER 8888
ENTRYPOINT ["python", "-m", "adserver"]
