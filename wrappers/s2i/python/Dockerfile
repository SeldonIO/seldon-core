ARG CONDA_VERSION
ARG BASE_IMAGE
ARG IMAGE_VERSION
FROM $BASE_IMAGE:$IMAGE_VERSION

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

# This is to install desired version of Python without updating conda version
ARG PYTHON_VERSION
RUN conda install --yes python=$PYTHON_VERSION conda=$CONDA_VERSION
RUN dnf install -y make automake gcc gcc-c++

# Upgrade pip version
RUN pip install pip==20.2 setuptools==46.1

RUN mkdir microservice
WORKDIR /microservice

COPY ./s2i/bin/ /s2i/bin

# keep install of seldon-core after the COPY to force re-build of layer
COPY requirements.txt ./requirements.txt
COPY version.txt /microservice/version.txt
COPY _python/python/licenses/license.txt .
RUN pip install -r requirements.txt

RUN mkdir -p /.conda && chmod a+rwx /.conda
RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

EXPOSE 5000
