import kafka
from multiprocessing import get_logger

logger = get_logger()


class KafkaDataIngestor:
    def __init__(self, bootstrap_servers, consumer_topic, producer_topic, group_id):
        """
        This includes the inputs from the parameters, which we could provide some by default
        """
        self.bootstrap_servers = bootstrap_servers
        self.consumer_topic = consumer_topic
        self.producer_topic = producer_topic
        self.group_id = group_id

    def connection(self):
        consumer = kafka.KafkaConsumer(
            self.consumer_topic,
            group_id=self.group_id,
            bootstrap_servers=self.bootstrap_servers,
        )
        producer = kafka.KafkaProducer(bootstrap_servers=self.bootstrap_servers)
        return {
            "consumer": consumer,
            "producer": producer,
            "producer_topic": self.producer_topic,
        }

    @staticmethod
    def fetch(self, connection):
        """
        Callects the single next message in the queue to process or
        waits until there is a new message. It leverages the yield
        functionality that the KafkaConsumer exposes.
        """
        return next(connection["consumer"])

    @staticmethod
    def process(seldon_client, connection, in_data):
        out_data = seldon_client.predict(data=in_data)
        return out_data

    @staticmethod
    def publish(out_data, in_data, connection):
        logger.info(f"Publishing with out_data=[{out_data}], in_data={in_data}")

        res = connection["producer"].send(connection["producer_topic"], in_data)

        logger.info(f"Publish result: {res}")
