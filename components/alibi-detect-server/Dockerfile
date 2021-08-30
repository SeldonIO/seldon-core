# TODO: Add to release script
FROM docker.io/seldonio/seldon-core-s2i-python37-ubi8:1.11.0-dev
LABEL name="Seldon Alibi Detect Server" \
      vendor="Seldon Technologies" \
      version="1.11.0-dev" \
      release="1" \
      summary="Alibi Detect Server for Seldon Core" \
      description="The Alibi Detect Server provides outlier, drift and adversarial detection services for Seldon Core"

# Install Rclone Binary to be present in the image
RUN yum install -y unzip
RUN wget https://downloads.rclone.org/v1.55.1/rclone-v1.55.1-linux-amd64.zip && \
    unzip rclone-v1.55.1-linux-amd64.zip && \
    mv rclone-v1.55.1-linux-amd64/rclone /usr/bin/rclone && \
    rm -rf rclone-v1.55.1-linux-amd64.zip rclone-v1.55.1-linux-amd64

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


RUN yum install -y mesa-libGL
RUN mv ./licenses /licenses
RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

ENTRYPOINT ["python", "-m", "adserver"]
