from confluent_kafka import Producer, Consumer, KafkaError, KafkaException, Message
import socket
import argparse
import sys
from tensorflow_serving.apis import predict_pb2
from google.protobuf.json_format import MessageToJson

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("command", type=str, choices=["produce","consume"])
    parser.add_argument("broker", type=str, help="Kafka Broker")
    parser.add_argument("topic", type=str, help="Kafka Topic")
    parser.add_argument("--file", type=str, help="file to read data from")
    parser.add_argument("--proto_name", type=str, help="proto type name", default="")
    args = parser.parse_args()
    conf = {'bootstrap.servers': args.broker,
        'client.id': socket.gethostname(),
        'group.id': "foo7",
        'auto.offset.reset': 'smallest'
        }

    if args.command == "produce":
        produce(conf, args.topic, args.file, args.proto_name)
    elif args.command == "consume":
        consume(conf,args.topic, args.proto_name)


def decode_proto(msg, proto_name):
    print(proto_name)
    if proto_name == "tensorflow.serving.PredictResponse":
        pr: predict_pb2.PredictResponse = predict_pb2.PredictResponse()
        pr.ParseFromString(msg.value())
        print(MessageToJson(pr))
    else:
        print("unknown proto received ", proto_name)

def consume(conf, topic, proto_name):
    consumer = Consumer(conf)
    running = True
    cnt = 0
    try:
        consumer.subscribe([topic])

        while running:
            msg: Message = consumer.poll(timeout=1.0)
            if msg is None: continue

            if msg.error():
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    # End of partition event
                    sys.stderr.write('%% %s [%d] reached end at offset %d\n' %
                                     (msg.topic(), msg.partition(), msg.offset()))
                elif msg.error():
                    raise KafkaException(msg.error())
            else:
                cnt += 1
                if proto_name == "":
                    print(msg.value())
                else:
                    decode_proto(msg, proto_name)
                print("Consumed ", cnt)
    finally:
        # Close down consumer to commit final offsets.
        consumer.close()

def shutdown():
    running = False

def produce(conf, topic, file, protoName):
    producer = Producer(conf)
    if protoName == "":
        produce_text(producer, topic,file)
    else:
        produce_proto(producer, topic, file, protoName)


def produce_proto(producer, topic, file, protoName):
    with open(file, "rb") as fp:
        cnt = 0
        szBytes = fp.read(4)
        while szBytes:
            cnt = cnt + 1
            sz = int.from_bytes(szBytes, byteorder='big')
            data = fp.read(sz)
            headers = {"proto-name":protoName}
            producer.produce(topic, value=data, headers=headers)
            producer.poll(0)
            if cnt % 100 == 0:
                print("Messages sent:",cnt)
                producer.flush()
            szBytes = fp.read(4)
    producer.flush()
    print("Final messages sent count:", cnt)

def produce_text(producer, topic, file):
    with open(file,"r") as fp:
        cnt = 0
        line = fp.readline()
        while line:
            cnt = cnt + 1
            producer.produce(topic, value=line)
            producer.poll(0)
            line = fp.readline()
            if cnt % 100 == 0:
                print("messages sent:",cnt)
                producer.flush()

    producer.flush()
    print("Final messages sent count:", cnt)



if __name__ == "__main__":
    main()