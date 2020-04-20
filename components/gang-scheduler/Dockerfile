FROM python:3.8-slim-buster

RUN mkdir /app
WORKDIR /app

COPY . /app/.
RUN cd /app/ && pip install -e .

CMD ["sh"]

