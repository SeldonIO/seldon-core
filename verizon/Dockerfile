FROM python:3.11-slim

USER root

ENV NEW_RELIC_CONFIG_FILE=/opt/mlserver/newrelic.ini

RUN apt update -y
RUN apt upgrade -y
RUN apt install build-essential g++ -y
RUN apt install libpcre2-dev swig net-tools lsof curl wget htop vim nano cron git -y


RUN pip install --upgrade mlserver==1.6.1 \
                          #pydantic==1.10.19 \
                          newrelic==9.0.0

RUN pip install --upgrade starlette==0.41.2


ENV MLSERVER_PATH=/opt/mlserver
RUN mkdir $MLSERVER_PATH
WORKDIR /opt/mlserver

RUN useradd -u 1000 -s /bin/bash mlserver -d $MLSERVER_PATH && \
    chown -R 1000:0 $MLSERVER_PATH && \
    chmod -R 776 $MLSERVER_PATH


#COPY newrelic.ini $NEW_RELIC_CONFIG_FILE

SHELL ["/bin/bash", "-c"]
USER root
COPY requirements.txt .
RUN pip uninstall -y tensorflow tensorboard numpy pandas typing_extensions
RUN pip install -r requirements.txt
RUN mkdir /opt/models
USER 1000
CMD newrelic-admin run-program mlserver start /opt/models