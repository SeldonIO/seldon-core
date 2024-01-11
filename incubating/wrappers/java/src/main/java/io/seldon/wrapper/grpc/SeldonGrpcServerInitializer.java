/**
 * *****************************************************************************
 * 
 * Copyright (c) 2024 Seldon Technologies Ltd.
 * 
 * Use of this software is governed BY
 * (1) the license included in the LICENSE file or
 * (2) if the license included in the LICENSE file is the Business Source License 1.1,
 * the Change License after the Change Date as each is defined in accordance with the LICENSE file.
 *
 * *****************************************************************************
 */
package io.seldon.wrapper.grpc;

import java.io.IOException;
import javax.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class SeldonGrpcServerInitializer {

  protected static Logger logger =
      LoggerFactory.getLogger(SeldonGrpcServerInitializer.class.getName());

  @Autowired SeldonGrpcServer server;

  @PostConstruct
  public void initialise() {
    try {
      server.runServer();
    } catch (InterruptedException e) {
      logger.error("Failed to start grc server", e);
    } catch (IOException e) {
      logger.error("Failed to start grc server", e);
    }
  }
}
