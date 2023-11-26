FROM seldonio/seldon-core-s2i-python3:0.18
COPY . /app
WORKDIR /app
RUN pip install -r requirements.txt
EXPOSE 5000

# Define environment variable
ENV MODEL_NAME iris
ENV SERVICE_TYPE MODEL
ENV PERSISTENCE 0

CMD exec seldon-core-microservice $MODEL_NAME --service-type $SERVICE_TYPE --persistence $PERSISTENCE --access-log