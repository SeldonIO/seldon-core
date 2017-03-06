package io.seldon.clustermanager.zk;

import org.apache.curator.RetryPolicy;
import org.apache.curator.framework.CuratorFramework;
import org.apache.curator.framework.CuratorFrameworkFactory;
import org.apache.curator.retry.ExponentialBackoffRetry;

import io.seldon.clustermanager.component.ZookeeperManager;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ZookeeperManagerImpl implements ZookeeperManager {

    private final static Logger logger = LoggerFactory.getLogger(ZookeeperManagerImpl.class);

    private CuratorFramework curator = null;

    public void init() throws Exception {
        logger.info("init");
        String zookeeperConnectionString = "localhost";
        RetryPolicy retryPolicy = new ExponentialBackoffRetry(1000, 3);
        curator = CuratorFrameworkFactory.newClient(zookeeperConnectionString, retryPolicy);
        curator.start();

        try {
            byte[] v = curator.getData().forPath("/"); // Check we can get the root node to see if all is working
            logger.info("Sucessfully passed root node data check");
        } catch (Exception e) {
            throw e;
        }
    }

    public void cleanup() throws Exception {
        logger.info("cleanup");
        if (curator != null) {
            curator.close();
        }
    }

}
