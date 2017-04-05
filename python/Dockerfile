FROM openjdk:8

RUN apt-get update -y
RUN apt-get install -y python-pip python-dev build-essential

COPY .

RUN python setup.py install