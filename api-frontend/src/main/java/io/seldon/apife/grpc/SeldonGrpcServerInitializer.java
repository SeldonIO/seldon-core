package io.seldon.apife.grpc;

import java.io.IOException;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class SeldonGrpcServerInitializer {

    protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServerInitializer.class.getName());
    
    @Autowired
    SeldonGrpcServer server;
    
    @PostConstruct
    public void initialise() {
        try
        {
            server.runServer();
        } catch (InterruptedException e) {
            logger.error("Failed to start grc server",e);
        } catch (IOException e) {
            logger.error("Failed to start grc server",e);
        }
    }
    
}
