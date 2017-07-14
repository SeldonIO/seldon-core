#FROM openjdk:8
FROM ubuntu:16.04

RUN apt-get update -y
RUN apt-get install -y python-pip python-dev build-essential
RUN apt-get install -y python-numpy python-scipy

COPY . /pyseldon
WORKDIR /pyseldon

RUN python setup.py install
