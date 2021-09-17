ARG VERSION
FROM seldonio/seldon-core-s2i-python37-ubi8:1.11.0
LABEL name="Seldon Alibi Wrapper" \
      vendor="Seldon Technologies" \
      version="1.11.0" \
      release="1" \
      summary="Alibi Explainer Wrapper for Seldon Core" \
      description="Allows Seldon Core inference models to run with a black box model explanation model from the Alibi:Explain project"

RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical

RUN mkdir /licenses 

# CVE https://github.com/SeldonIO/seldon-core/issues/2960
RUN yum remove -y nodejs httpd

RUN pip install \
    --upgrade pip \
    --upgrade setuptools \
    --upgrade wheel

# https://github.com/slundberg/shap/pull/1802#issuecomment-846299228
# https://github.com/SeldonIO/alibi/issues/466
RUN pip install numpy==1.19 numba==0.53.1

COPY setup.py setup.py
COPY alibiexplainer alibiexplainer
COPY README.md README.md

RUN pip install . --no-binary protobuf

# Add licences
RUN pip install pip-licenses
RUN mkdir ./licenses && pip-licenses --from=mixed --format=csv --output-file=./licenses/license_info.csv && \
    pip-licenses --from=mixed --format=plain-vertical --with-license-file --no-license-path --output-file=./licenses/license.txt
RUN cp ./licenses/* /licenses

# Install python spacy model to avoid issues in airgapped envs
RUN python -m spacy download en_core_web_md

ENTRYPOINT ["python", "-m", "alibiexplainer"]
