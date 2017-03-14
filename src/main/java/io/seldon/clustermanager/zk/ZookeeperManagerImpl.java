package io.seldon.clustermanager.zk;

import org.apache.curator.RetryPolicy;
import org.apache.curator.framework.CuratorFramework;
import org.apache.curator.framework.CuratorFrameworkFactory;
import org.apache.curator.retry.ExponentialBackoffRetry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class ZookeeperManagerImpl implements ZookeeperManager {

    private final static Logger logger = LoggerFactory.getLogger(ZookeeperManagerImpl.class);
    private final static String SELDON_CLUSTER_MANAGER_ZK_SERVERS_KEY = "SELDON_CLUSTER_MANAGER_ZK_SERVERS";

    private CuratorFramework curator = null;

    public void init() throws Exception {
        logger.info("init");

        String zookeeperConnectionString = "UNDEFINED";
        { // setup the zookeeperConnectionString using the env vars
            zookeeperConnectionString = System.getenv().get(SELDON_CLUSTER_MANAGER_ZK_SERVERS_KEY);
            if (zookeeperConnectionString == null) {
                logger.error(String.format("FAILED to find env var [%s]", SELDON_CLUSTER_MANAGER_ZK_SERVERS_KEY));
                zookeeperConnectionString = "localhost:2181";
            }
            logger.info(String.format("Setting zookeeper connection string as [%s]", zookeeperConnectionString));
        }

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

    @Override
    public void persistSeldonDeployment(DeploymentDef deploymentDef) throws Exception {
        String json = ProtoBufUtils.toJson(deploymentDef, true);
        byte[] json_bytes = json.getBytes("UTF-8");

        final String seldonDeploymentId = deploymentDef.getId();
        String deployment_node_path = String.format("/deployments/%s", seldonDeploymentId);

        if (curator.checkExists().forPath(deployment_node_path) == null) {
            // Create node
            curator.create().creatingParentsIfNeeded().forPath(deployment_node_path, json_bytes);
            logger.debug(String.format("[CREATE] [%s] [%s]", deployment_node_path, json));
        } else {
            // Update node
            curator.setData().forPath(deployment_node_path, json_bytes);
            logger.debug(String.format("[UPDATE] [%s] [%s]", deployment_node_path, json));
        }

    }

    @Override
    public void deleteSeldonDeployment(DeploymentDef deploymentDef) throws Exception {
        final String seldonDeploymentId = deploymentDef.getId();
        String deployment_node_path = String.format("/deployments/%s", seldonDeploymentId);

        curator.delete().deletingChildrenIfNeeded().forPath(deployment_node_path);
        logger.debug(String.format("[DELETE] [%s]", deployment_node_path));
    }

}
