
# Seldon Batch Processing Elasticsearch Data Source

This example will walk you through the steps for you to leverage the data ingestor batch functionality in Seldon Core to build your own custom Elasticsearch Data Ingestor which will consume from a Kafka Topic, send the data to Seldon and then publish the output into another Kafka topic. The diagram below provides some intuition on what we'll be deploying:

[](./batch-seldon-elk-cluster.png)

In this example we will:
1. Create an ELK Data Ingestor
2. Run our ELK Data Ingestor with docker-compose
3. Test the Data Ingestor in Kubernetes Cluster with test model

## 1) Create ELK Data Ingestor

To get started we want to create a simple data ingestor that consumes and publishes from ELK.

For this, we will create the file `ingestor/ElasticDataIngestor.py` with the following contents:

```python
from elasticsearch import Elasticsearch
from elasticsearch_dsl import Search

class ElasticDataIngestor:

    def __init__(
            self,
            host,
            port,
            index,
            data_ids):
        """
        This includes the inputs from the parameters, which we could provide some by default
        """
        self.host = host
        self.port = port
        self.index = index
        self.data_ids = data_ids

    def connection(self):
        es = Elasticsearch(hosts=[self.host], port=self.port)
        connection = {
            "client": es,
            "index": self.index
        }
        return connection

    @staticmethod
    def fetch(self, connection):
        search = (Search(using=connection["client"], index=connection["index"])
                  .query("match", logger_name=self.celery_task.task_id)
                  .sort("@timestamp")
                  .source(['message'])
                  .params(preserve_order=True))

        result = search.scan()
        in_data = [log for log in result]
        return False, in_data

    @staticmethod
    def process(seldon_client, connection, in_data):
        out_data = seldon_client.predict(data=in_data)
        return out_data

    @staticmethod
    def publish(out_data, in_data, connection):
        res = connection["client"].index(
            connection["index"], body=out_data)
```

## 2) Containerise the data ingestor 

In order to containerise the data ingestor we need to create a new file `ingestor/s2i/environment` with the following contents:

```bash
MODEL_NAME=ElasticDataIngestor
API_TYPE=BATCH
```

With this now we just have to containerise our wrapper using the Seldon CLI tools by running:

```
s2i build ingestor/. seldonio/seldon-core-s2i-python37:0.18-SNAPSHOT elastic_data_ingestor:0.1
```

## 3. Test the Data Ingestor in Kubernetes Cluster with test model

In order to test the model locally we need to perform the following steps:


TODO

