package io.seldon.engine.grpc;

import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.SeldonGrpc;

public class SeldonClientExample {
	protected static Logger logger = LoggerFactory.getLogger(SeldonClientExample.class.getName());
	
	 private final ManagedChannel channel;
	  private final SeldonGrpc.SeldonBlockingStub blockingStub;
	  private final SeldonGrpc.SeldonStub asyncStub;

	  /** Construct client for accessing RouteGuide server at {@code host:port}. */
	  public SeldonClientExample(String host, int port) {
	    this(ManagedChannelBuilder.forAddress(host, port).usePlaintext(true));
	  }

	  /** Construct client for accessing RouteGuide server using the existing channel. */
	  public SeldonClientExample(ManagedChannelBuilder<?> channelBuilder) {
	    channel = channelBuilder.build();
	    blockingStub = SeldonGrpc.newBlockingStub(channel);
	    asyncStub = SeldonGrpc.newStub(channel);
	  }

	  public void shutdown() throws InterruptedException {
		    channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
		}
	  
	  public void predict() throws InvalidProtocolBufferException
	  {
		  SeldonMessage request = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(1.0).addShape(1))).build();
		  
		  SeldonMessage response = blockingStub.predict(request);
		  
		  logger.info(ProtoBufUtils.toJson(response));
	  }
	  
	  /** Issues several different requests and then exits. 
	 * @throws InvalidProtocolBufferException */
	  public static void main(String[] args) throws InterruptedException, InvalidProtocolBufferException {

	    SeldonClientExample client = new SeldonClientExample("localhost", SeldonGrpcServer.SERVER_PORT);
	    try {
	    	
	    	client.predict();
	    
	    } finally {
	      client.shutdown();
	    }
	  }
	  
	
}
