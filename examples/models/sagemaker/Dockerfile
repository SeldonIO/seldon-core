FROM sklearn-base:0.20.0-cpu-py3

ADD requirements.txt .

RUN pip install -r requirements.txt

COPY model /opt/ml/model

COPY code /opt/ml/code

COPY serving.py /serving.py

