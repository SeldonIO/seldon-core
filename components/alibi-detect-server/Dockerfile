FROM registry.access.redhat.com/ubi8/python-36
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.2.4" \
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
RUN mv ./licenses /licenses
USER default

ENTRYPOINT ["python", "-m", "adserver"]
