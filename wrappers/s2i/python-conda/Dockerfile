FROM continuumio/miniconda3:4.7.12

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

RUN mkdir microservice
WORKDIR /microservice

COPY ./s2i/bin/ /s2i/bin

COPY requirements.txt ./requirements.txt
RUN pip install -r requirements.txt

EXPOSE 5000
