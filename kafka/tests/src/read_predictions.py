from __future__ import print_function
import os
import sys, getopt, argparse
import logging
import json
from kafka import KafkaConsumer
from kafka import TopicPartition
import prediction_pb2 as ppb

if __name__ == '__main__':
    import logging
    logger = logging.getLogger()
    logging.basicConfig(format='%(asctime)s : %(levelname)s : %(name)s : %(message)s', level=logging.DEBUG)
    logger.setLevel(logging.INFO)

    parser = argparse.ArgumentParser(prog='read_predictions')
    parser.add_argument('--kafka', help='kafka endpoint', default="localhost:9093")
    parser.add_argument('--topic', help='kafka topic', required=True)

    args = parser.parse_args()
    opts = vars(args)

    consumer = KafkaConsumer(client_id="py-kafka",group_id=None,bootstrap_servers=args.kafka)
    partition = TopicPartition(args.topic, 0)
    consumer.assign([partition])
    consumer.seek(partition, 0)
    for msg in consumer:
        print(msg)
        message = ppb.PredictionRequestResponseDef()
        message.ParseFromString(msg.value)
        print(message)
