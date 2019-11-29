FROM python:%PYTHON_VERSION%

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

RUN apt-get update -y
RUN apt-get install -y python-pip python-dev build-essential

RUN mkdir microservice
WORKDIR /microservice

COPY ./s2i/bin/ /s2i/bin

# keep install of seldon-core after the COPY to force re-build of layer
RUN pip install seldon-core

EXPOSE 5000
