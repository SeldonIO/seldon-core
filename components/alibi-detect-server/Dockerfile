FROM python:3.7

RUN apt-get update \
    && apt-get install -y --no-install-recommends git \
    && apt-get purge -y --auto-remove \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

ADD requirements_server.txt .

RUN pip install -r requirements_server.txt

# Fix cloudevents bug: https://github.com/cloudevents/sdk-python/issues/24
RUN git clone --branch 24-extensions https://github.com/ryandawsonuk/sdk-python.git && \
    cd sdk-python && \
    pip install -e .

COPY adserver adserver
COPY setup.py .

RUN pip install -e .

ENTRYPOINT ["python", "-m", "adserver"]
