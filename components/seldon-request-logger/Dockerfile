FROM python:3.7-slim
COPY . /app
WORKDIR /app
RUN pip install -r requirements.txt
RUN pip install gunicorn
EXPOSE 8080
ENTRYPOINT ["python"]
CMD ["app/default_logger.py"]