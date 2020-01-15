FROM seldonio/seldon-core-s2i-python37:0.14

COPY requirements-dev.txt /microservice/
RUN pip install --upgrade pip setuptools && \
    pip install -r requirements-dev.txt
