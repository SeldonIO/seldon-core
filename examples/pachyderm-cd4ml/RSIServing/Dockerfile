FROM python:3.8-slim
RUN pip install pandas==1.1.4 statsmodels==0.12.1 seldon-core==1.5.0 joblib==1.0.0 minio==7.0.0
WORKDIR /app
ADD RSIModel.py /app/RSIModel.py
EXPOSE 5000
CMD exec seldon-core-microservice RSIModel --service-type MODEL --persistence 0