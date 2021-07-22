FROM registry.access.redhat.com/ubi8/python-38
LABEL name="Seldon Request Logger" \
      vendor="Seldon Technologies" \
      version="1.10.0-dev" \
      release="1" \
      summary="The payload logger for Seldon Core" \
      description="The Seldon Payload Logger allows request and response payloads from a Seldon Core inference graph to be processed and sent to an ELK endpoint"

COPY requirements.txt requirements.txt
RUN pip install -r requirements.txt
RUN pip install gunicorn

# Add licences
RUN pip install pip-licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
USER root
RUN mv ./licenses /licenses
RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

# CVE https://github.com/SeldonIO/seldon-core/issues/2960
RUN yum remove -y nodejs httpd

USER default

ENV APP_ENV=production \
  PATH="/home/appuser/.local/bin:${PATH}" \
  PORT=8080 \
  GUNICORN_WORKERS=3 \
  GUNICORN_THREADS=8

COPY app app
WORKDIR app

EXPOSE $PORT

CMD gunicorn --bind 0.0.0.0:$PORT --workers $GUNICORN_WORKERS --threads $GUNICORN_THREADS default_logger:app

