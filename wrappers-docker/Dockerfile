FROM python:2.7.14-jessie

COPY _wrappers /wrappers

# install docker
RUN \
    apt-get update && \
    apt-get install -y \
         apt-transport-https \
         ca-certificates \
         curl \
         gnupg2 \
         software-properties-common && \
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add - && \
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
       $(lsb_release -cs) \
       stable" && \
    apt-get update && \
    apt-get install -y docker-ce



RUN python -m pip install grpcio-tools==1.1.3
RUN cd /wrappers && make build_protos

# wrapper python dependency
RUN pip install jinja2

# deps to get tester.py to work
RUN pip install numpy requests redis flask

WORKDIR /wrappers/python

ENTRYPOINT ["python","wrap_model.py"]
CMD []