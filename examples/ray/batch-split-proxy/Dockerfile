FROM seldonio/seldon-core-s2i-python37:1.5.0-dev

# Add Model
COPY training/outputs/pytorch_model.bin /microservice/pytorch_model.bin

# Install requirements
COPY requirements.txt /microservice
RUN pip install -r requirements.txt && rm -r ~/.cache/pip

RUN pip install ray[tune]

# Add file that will download Roberta cache inside container
COPY import_for_cache_download.py /tmp
RUN python3 /tmp/import_for_cache_download.py

# Add Seldon Model
ENV MODEL_NAME Model
ENV API_TYPE REST

COPY Model.py /microservice/
CMD seldon-core-microservice $MODEL_NAME $API_TYPE
