kafka-console-producer.sh --broker-list `cat auth/broker.ip`:9093 --topic test --producer.config config.properties
kafka-console-consumer.sh --bootstrap-server `cat auth/broker.ip`:9093 --topic test --consumer.config config.properties --from-beginning
