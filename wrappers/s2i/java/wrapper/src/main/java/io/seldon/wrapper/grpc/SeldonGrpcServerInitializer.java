/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.wrapper.grpc;

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
