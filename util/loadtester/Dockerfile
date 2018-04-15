FROM ubuntu:16.04

ENV HOME /root
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -y
RUN apt-get install gcc g++ python-gevent libzmq3-dev python-dev python-pip git -y

COPY requirements.txt /tmp/
RUN pip install -r /tmp/requirements.txt

RUN git clone https://github.com/locustio/locust && \
    cd locust && \
    python setup.py install

ENV SELDON_HOME /home/seldon
ADD ./scripts $SELDON_HOME/scripts

# Define default command.
CMD ["bash"]

