package io.seldon.clustermanager.zk;

import io.seldon.clustermanager.component.ZookeeperManager;

public class ZookeeperManagerImpl implements ZookeeperManager {

    public void init() throws Exception {
        System.out.println("ZookeeperManager init");
    }

    public void cleanup() throws Exception {
        System.out.println("ZookeeperManager cleanup");
    }

}
