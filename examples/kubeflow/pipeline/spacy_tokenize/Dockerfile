FROM python:3.7-slim
COPY . /app
WORKDIR /app
RUN pip install -r requirements.txt
RUN python -m spacy download en_core_web_sm
