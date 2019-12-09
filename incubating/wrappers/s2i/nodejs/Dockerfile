FROM node:latest

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

RUN mkdir microservice
WORKDIR /microservice

COPY *microservice.js /microservice/

COPY package.json /microservice/

COPY prediction_grpc_pb.js /microservice/

COPY prediction_pb.js /microservice/

RUN npm install

COPY ./s2i/bin/ /s2i/bin

EXPOSE 5000
