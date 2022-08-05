ARG VERSION
ARG BASE_IMAGE
FROM ${BASE_IMAGE}:${VERSION} as base

ARG VERSION
LABEL name="Seldon Alibi Wrapper" \
      vendor="Seldon Technologies" \
      version="1.14.1" \
      release="1" \
      summary="Alibi Explainer Wrapper for Seldon Core" \
      description="Allows Seldon Core inference models to run with a black box model explanation model from the Alibi:Explain project"

FROM base as builder
ARG PYTHON_VERSION
ARG CONDA_VERSION

# Install Python / Conda
RUN conda install --yes python=${PYTHON_VERSION} conda=${CONDA_VERSION}
RUN pip install pip==21.2.4 setuptools==58.1.0
RUN dnf install -y make automake gcc gcc-c++

# Make home dir
RUN mkdir microservice
WORKDIR /microservice

# Install Poetry
ENV POETRY_HOME /microservice/.poetry
RUN curl -sSL \
    https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py \
    | python3

ENV PATH "$POETRY_HOME/bin:$PATH"
ENV POETRY_VIRTUALENVS_CREATE false

# Install the server
COPY poetry.lock pyproject.toml ./
RUN poetry install && rm ~/.cache/pip -rf


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


FROM base as final
WORKDIR /microservice

RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical
# CVE https://github.com/SeldonIO/seldon-core/issues/2960
RUN yum remove -y nodejs httpd

COPY --from=builder /microservice /microservice
COPY --from=builder /licenses /licenses
COPY --from=builder /opt/conda /opt/conda

ENTRYPOINT ["python", "-m", "alibiexplainer"]
