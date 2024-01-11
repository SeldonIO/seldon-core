ARG BASE_IMAGE
ARG VERSION
FROM $BASE_IMAGE:$VERSION

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

# We are setting default image user as 8888 but s2i build may require root
LABEL io.openshift.s2i.assemble-user="0"

# This is to install desired version of Python without updating conda version
# Note: ARGs must be after FROM statement , see source:
# https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact

ARG PYTHON_VERSION
ARG CONDA_VERSION

# MPL based licenses
RUN wget -O certifi-python-certifi.tar.gz https://github.com/certifi/python-certifi/archive/master.tar.gz

RUN microdnf update -y && \
    microdnf install -y make automake gcc gcc-c++ && \
    conda install -c conda-forge --yes python=$PYTHON_VERSION conda=$CONDA_VERSION && \
    pip install --upgrade pip setuptools wheel

RUN mkdir microservice
WORKDIR /microservice

COPY ./s2i/bin/ /s2i/bin

# keep install of seldon-core after the COPY to force re-build of layer
COPY version.txt /microservice/version.txt
COPY _python/python/licenses/license.txt .
COPY _python /microservice
RUN cd /microservice/python && pip install .

# Default user is 8888 so changing for permissiosn to modify (needed in mlflow
# server).
# Clean Conda's cache to avoid cache file conflicts between user IDs:
# https://github.com/conda/conda/issues/7227
RUN mkdir -p /.conda /.cache/conda && \
    chmod a+rwx /.conda /.cache/conda && \
    chown -R 8888 /microservice  && \
    conda clean --yes --index-cache

USER 8888
EXPOSE 5000
