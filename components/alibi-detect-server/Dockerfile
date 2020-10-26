# TODO: Add to release script 
FROM docker.io/seldonio/seldon-core-s2i-python37-ubi8:1.3.0-dev
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.4.0-dev_20201026094752" \
      release="1" \
      summary="Alibi Detect Server for Seldon Core" \
      description="The Alibi Detect Server provides outlier, drift and adversarial detection services for Seldon Core"

ADD requirements_server.txt .

RUN pip install pip -U

RUN pip install -r requirements_server.txt

# Fix cloudevents bug: https://github.com/cloudevents/sdk-python/issues/24
RUN git clone --branch 24-extensions https://github.com/ryandawsonuk/sdk-python.git && \
    cd sdk-python && \
    pip install -e .

COPY adserver adserver
COPY setup.py .

RUN pip install -e .

# Add licences
RUN pip install pip-licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
USER root
RUN yum install -y mesa-libGL
RUN mv ./licenses /licenses
USER default

ENTRYPOINT ["python", "-m", "adserver"]
