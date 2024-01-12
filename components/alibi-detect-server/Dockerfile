ARG VERSION
ARG BASE_IMAGE
FROM ${BASE_IMAGE}:${VERSION} as base

ARG VERSION
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.19.0-dev" \
      release="1" \
      summary="Alibi Detect Server for Seldon Core" \
      description="The Alibi Detect Server provides outlier, drift and adversarial detection services for Seldon Core"

FROM base as builder

RUN microdnf update -y && \
    microdnf install -y \
      unzip \
      make \
      automake \
      gcc \
      gcc-c++

# Install Rclone Binary to be present in the image
RUN wget https://downloads.rclone.org/v1.64.2/rclone-v1.64.2-linux-amd64.zip && \
    unzip rclone-v1.64.2-linux-amd64.zip && \
    mv rclone-v1.64.2-linux-amd64/rclone /usr/bin/rclone && \
    rm -rf rclone-v1.64.2-linux-amd64.zip rclone-v1.64.2-linux-amd64

# Note that we need to force Conda to use the system's std-c++ library, as
# otherwise PyStan won't compile with Conda's older stdc++ library:
# https://github.com/stan-dev/pystan/issues/294#issuecomment-870711100
RUN pip install --upgrade pip setuptools wheel && \
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
## NOTE: Removing explicitly requirements.txt file from subdeps test
## dependencies causing false positives in Snyk.
COPY poetry.lock pyproject.toml ./
COPY _seldon_core ./_seldon_core
RUN poetry install && \
    rm ~/.cache/pip -rf && \
    rm -f /opt/conda/lib/python3.8/site-packages/gslib/vendored/boto/requirements.txt \
          /opt/conda/lib/python3.8/site-packages/gslib/vendored/oauth2client/docs/requirements.txt \
          /opt/conda/lib/python3.8/site-packages/tests/conda_env/support/requirements.txt

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

ENV RCLONE_CONFIG_GS_TYPE="google cloud storage" \
    RCLONE_CONFIG_GS_ANONYMOUS="true" \
    DRIFT_ARTIFACTS_DIR="/mnt/artifacts/"

# mesa-libGL: this is to avoid "ImportError: libGL.so.1" from opencv
RUN microdnf install -y mesa-libGL
RUN microdnf update -y

COPY --from=builder /microservice /microservice
COPY --from=builder /opt/conda /opt/conda
COPY --from=builder /usr/bin/rclone /usr/bin/rclone
COPY --from=builder /licenses /licenses

# This is to have writable numba and keops cache directories
ENV NUMBA_CACHE_DIR /tmp/numba-cache
RUN mkdir /.cache && \
    chown -R 8888:0 /.cache && \
    chmod -R 776 /.cache

RUN mkdir -p /mnt/artifacts && \
    chown -R 8888:0 /mnt/artifacts && \
    chmod -R 776 /mnt/artifacts

USER 8888
ENTRYPOINT ["python", "-m", "adserver"]
