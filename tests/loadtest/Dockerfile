FROM continuumio/anaconda

ENV HOME /root
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -y
RUN apt-get install gcc -y

RUN pip install locustio

ENV SELDON_HOME /home/seldon
ADD ./scripts $SELDON_HOME/scripts

# Define default command.
CMD ["bash"]

