
```python
from elasticsearch import Elasticsearch
from elasticsearch_dsl import Search

class ElasticDataIngestor:

    def __init__(
            self,
            host,
            port,
            index):
        """
        This includes the inputs from the parameters, which we could provide some by default
        """
        self.data_ids = data_ids
        self.host = host
        self.port = port
        self.index = index

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
        out_data = seldon_client.predict(data)
        return out_data

    @staticmethod
    def publish(out_data, in_data, connection):
        client.publish(data, output)

```

