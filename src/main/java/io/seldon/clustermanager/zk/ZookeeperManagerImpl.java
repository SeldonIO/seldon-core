package io.seldon.clustermanager.zk;

import org.apache.curator.RetryPolicy;
import org.apache.curator.framework.CuratorFramework;
import org.apache.curator.framework.CuratorFrameworkFactory;
import org.apache.curator.retry.ExponentialBackoffRetry;

import io.seldon.clustermanager.component.ZookeeperManager;

public class ZookeeperManagerImpl implements ZookeeperManager {

    private CuratorFramework curator = null;

    public void init() throws Exception {
        System.out.println("ZookeeperManager init");
        String zookeeperConnectionString = "localhost";
        RetryPolicy retryPolicy = new ExponentialBackoffRetry(1000, 3);
        curator = CuratorFrameworkFactory.newClient(zookeeperConnectionString, retryPolicy);
        curator.start();

        try {
            byte[] v = curator.getData().forPath("/"); // Check we can get the root node to see if all is working
            System.out.println("Sucessfully checked node '/' from zookeeper");
        } catch (Exception e) {
            throw e;
        }
    }

    public void cleanup() throws Exception {
        System.out.println("ZookeeperManager cleanup");
        if (curator != null) {
            curator.close();
        }
    }

}
