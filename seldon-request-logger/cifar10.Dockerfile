FROM python:3.7-slim
COPY . /cifar
WORKDIR /cifar
RUN pip install -r requirements.txt
RUN pip install gunicorn
EXPOSE 8080
ENTRYPOINT ["python"]
CMD ["app/app.py"]