ARG VERSION
ARG BASE_IMAGE
FROM ${BASE_IMAGE}:${VERSION} AS base

ARG VERSION
LABEL name="Seldon Alibi Wrapper" \
      vendor="Seldon Technologies" \
      version="1.19.0-dev" \
      release="1" \
      summary="Alibi Explainer Wrapper for Seldon Core" \
      description="Allows Seldon Core inference models to run with a black box model explanation model from the Alibi:Explain project"

FROM base AS builder

USER root

# Install Python / Conda
RUN microdnf update -y && \
    pip install --upgrade pip setuptools wheel && \
    microdnf install -y git make automake gcc gcc-c++

# Make home dir
RUN mkdir microservice
WORKDIR /microservice

# Install Poetry
ENV POETRY_HOME=/microservice/.poetry
RUN /opt/conda/bin/pip install --no-cache-dir "poetry==2.1.2"

# Replace vulnerable pip wheel embedded inside virtualenv(CVE-2025-8869)
# Replace conda's python ensurepip with correct version
RUN PIP_WHEEL_DIR=/opt/conda/lib/python3.12/ensurepip/_bundled && \
    /opt/conda/bin/pip wheel "pip==25.3.0" "setuptools>=75.0.0,<76.0.0" "wheel>=0.44.0,<0.45.0" --wheel-dir $PIP_WHEEL_DIR && \
    rm -f $PIP_WHEEL_DIR/pip-25.0.1-py3-none-any.whl

ENV PATH="$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE=false

# Install the dependencies only
COPY poetry.lock pyproject.toml ./
RUN poetry install --no-root

# Add licences
RUN mkdir /licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
RUN cp ./licenses/* /licenses

# Install python spacy model to avoid issues in airgapped envs
RUN python -m spacy download en_core_web_md

# Copy rest of the package
COPY alibiexplainer alibiexplainer
COPY README.md README.md

# Install the project code
RUN poetry install

FROM base AS final
WORKDIR /microservice

RUN microdnf -y update

COPY --from=builder /microservice /microservice
COPY --from=builder /licenses /licenses
COPY --from=builder /opt/conda /opt/conda

USER 8888
ENTRYPOINT ["python", "-m", "alibiexplainer"]
