FROM registry.access.redhat.com/ubi8/python-36
LABEL name="Seldon Alibi Wrapper" \
      vendor="Seldon Technologies" \
      version="1.2.4" \
      release="1" \
      summary="Alibi Explainer Wrapper for Seldon Core" \
      description="Allows Seldon Core inference models to run with a black box model explanation model from the Alibi:Explain project"


COPY kfserving/python/alibiexplainer alibiexplainer
COPY kfserving/python/kfserving kfserving
COPY kfserving/python/third_party third_party

USER root
RUN pip install --upgrade pip && pip install -e ./kfserving
RUN git clone https://github.com/SeldonIO/alibi.git && \
    cd alibi && \
    pip install .
RUN pip install -e ./alibiexplainer

RUN chmod -R a+rwx /opt/app-root/lib/python3.6

# Add licences
RUN pip install pip-licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
RUN mv ./licenses /licenses
USER default

ENTRYPOINT ["python", "-m", "alibiexplainer"]
