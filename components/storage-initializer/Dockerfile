FROM registry.access.redhat.com/ubi8/python-36
LABEL name="Storage Initializer" \
      vendor="Seldon Technologies" \
      version="1.9.0" \
      release="1" \
      summary="Storage Initializer for Seldon Core" \
      description="Allows Seldon Core to download artifacts from cloud and local storage to a local volume"

USER root
COPY ./kfserving/python/kfserving kfserving
RUN pip install --upgrade pip && pip install ./kfserving

COPY ./kfserving/python/storage-initializer storage-initializer
COPY ./kfserving/python/third_party third_party

# Add licences
RUN pip install pip-licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
RUN mv ./licenses /licenses
RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

# CVE https://github.com/SeldonIO/seldon-core/issues/2960
RUN yum remove -y nodejs httpd

USER default

ENTRYPOINT ["./storage-initializer/scripts/initializer-entrypoint"]
