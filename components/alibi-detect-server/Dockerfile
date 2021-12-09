ARG VERSION
ARG BASE_IMAGE
FROM ${BASE_IMAGE}:${VERSION} as base

ARG VERSION
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.12.0" \
      release="1" \
      summary="Alibi Detect Server for Seldon Core" \
      description="The Alibi Detect Server provides outlier, drift and adversarial detection services for Seldon Core"

FROM base as builder
ARG PYTHON_VERSION
ARG CONDA_VERSION

# Install Rclone Binary to be present in the image
RUN yum install -y unzip
RUN wget https://downloads.rclone.org/v1.55.1/rclone-v1.55.1-linux-amd64.zip && \
    unzip rclone-v1.55.1-linux-amd64.zip && \
    mv rclone-v1.55.1-linux-amd64/rclone /usr/bin/rclone && \
    rm -rf rclone-v1.55.1-linux-amd64.zip rclone-v1.55.1-linux-amd64


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

# this is to avoid "ImportError: libGL.so.1" from opencv
RUN yum install -y mesa-libGL
RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical
# CVE https://github.com/SeldonIO/seldon-core/issues/2960
RUN yum remove -y nodejs httpd

COPY --from=builder /microservice /microservice
COPY --from=builder /opt/conda /opt/conda
COPY --from=builder /usr/bin/rclone /usr/bin/rclone
COPY --from=builder /licenses /licenses

ENTRYPOINT ["python", "-m", "adserver"]
