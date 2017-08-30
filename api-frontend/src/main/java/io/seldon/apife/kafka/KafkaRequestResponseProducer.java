package io.seldon.apife.kafka;

import java.util.Properties;
import java.util.concurrent.ExecutionException;

import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.common.serialization.StringSerializer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionRequestResponseDef;

@Component
public class KafkaRequestResponseProducer {
	private KafkaProducer<String, PredictionRequestResponseDef > producer;
	private final static String topic = "predictions";
	
	private static Logger logger = LoggerFactory.getLogger(KafkaRequestResponseProducer.class.getName());
    final private static String ENV_VAR_SELDON_KAFKA_SERVER = "SELDON_ENGINE_KAFKA_SERVER";

	public KafkaRequestResponseProducer() 
	{
		String kafkaHostPort = System.getenv(ENV_VAR_SELDON_KAFKA_SERVER);
        logger.info(String.format("using %s[%s]", ENV_VAR_SELDON_KAFKA_SERVER, kafkaHostPort));
        if (kafkaHostPort == null) {
            logger.warn("*WARNING* SELDON_KAFKA_SERVER environment variable not set!");
            kafkaHostPort = "localhost:9093";
        }
        logger.info("Starting kafka client with server "+kafkaHostPort);
	    Properties props = new Properties();
	    props.put("bootstrap.servers", kafkaHostPort);
	    props.put("client.id", "RequestResponseProducer");
	    producer = new KafkaProducer<>(props, new StringSerializer(), new PredictionRequestResponseSerializer());
	}
	
	public void send(String clientId, PredictionRequestResponseDef data)
	{
		 try {
			producer.send(new ProducerRecord<>(clientId,
			         data.getResponse().getMeta().getPuid(),
			         data)).get();
		} catch (InterruptedException | ExecutionException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
	}
	
	public void createTopic()
	{

	}
}


