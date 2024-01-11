#!/usr/bin/env ipython
import time
#!/usr/bin/env python
# coding: utf-8

#
# # Scale Seldon Deployments based on Prometheus Metrics.
# This notebook shows how you can scale Seldon Deployments based on Prometheus metrics via KEDA.
#
# [KEDA](https://keda.sh/) is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed.
#
# With the support of KEDA in Seldon, you can scale your seldon deployments with any scalers listed [here](https://keda.sh/docs/2.0/scalers/).
# In this example we will scale the seldon deployment with Prometheus metrics as an example.

# ## Install Seldon Core
#
# Install Seldon Core as described in [docs](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html)
#
# Make sure add `--set keda.enabled=true`

# ## Install Prometheus
#



# In[55]:


get_ipython().system('kubectl create namespace seldon-monitoring')
get_ipython().system('helm upgrade --install seldon-monitoring kube-prometheus     --version 8.3.2     --set fullnameOverride=seldon-monitoring     --namespace seldon-monitoring     --repo https://charts.bitnami.com/bitnami     --wait')





time.sleep(10)

# In[56]:


get_ipython().system('kubectl rollout status -n seldon-monitoring statefulsets/prometheus-seldon-monitoring-prometheus')


time.sleep(5)




# In[57]:


get_ipython().system('cat pod-monitor.yaml')






# In[58]:


get_ipython().system('kubectl apply -f pod-monitor.yaml')




# ## Install KEDA
#
# Follow the [docs for KEDA](https://keda.sh/docs/) to install.

# ## Create model with KEDA

# To create a model with KEDA autoscaling you just need to add a KEDA spec referring in the Deployment, e.g.:
# ```yaml
# kedaSpec:
#   pollingInterval: 15                                # Optional. Default: 30 seconds
#   minReplicaCount: 1                                 # Optional. Default: 0
#   maxReplicaCount: 5                                 # Optional. Default: 100
#   triggers:
#   - type: prometheus
#           metadata:
#             # Required
#             serverAddress: http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090
#             metricName: access_frequency
#             threshold: '10'
#             query: rate(seldon_api_executor_client_requests_seconds_count{model_name="classifier"}[1m])
# ```
# The full SeldonDeployment spec is shown below.



# In[59]:


VERSION = get_ipython().getoutput('cat ../../version.txt')
VERSION = VERSION[0]
VERSION






# In[60]:


get_ipython().run_cell_magic('writefile', 'model_with_keda_prom.yaml', 'apiVersion: machinelearning.seldon.io/v1\nkind: SeldonDeployment\nmetadata:\n  name: seldon-model\nspec:\n  name: test-deployment\n  predictors:\n  - componentSpecs:\n    - spec:\n        containers:\n        - image: seldonio/mock_classifier:1.16.0-dev\n          imagePullPolicy: IfNotPresent\n          name: classifier\n          resources:\n            requests:\n              cpu: \'0.5\'\n      kedaSpec:\n        pollingInterval: 15                                # Optional. Default: 30 seconds\n        minReplicaCount: 1                                 # Optional. Default: 0\n        maxReplicaCount: 5                                 # Optional. Default: 100\n        triggers:\n        - type: prometheus\n          metadata:\n            # Required\n            serverAddress: http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090\n            metricName: access_frequency\n            threshold: \'10\'\n            query: rate(seldon_api_executor_client_requests_seconds_count{model_name="classifier"}[1m])\n    graph:\n      children: []\n      endpoint:\n        type: REST\n      name: classifier\n      type: MODEL\n    name: example')






# In[62]:


get_ipython().system('kubectl create -f model_with_keda_prom.yaml')





time.sleep(10)

# In[63]:


get_ipython().system("kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')")


time.sleep(5)


# ## Create Load

# We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.



# In[64]:


get_ipython().system("kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust")
get_ipython().system("kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[1].metadata.name}') role=locust")




# Before add loads to the model, there is only one replica



# In[65]:


get_ipython().system('kubectl get deployment seldon-model-example-0-classifier')






# In[66]:


get_ipython().system('helm install seldon-core-loadtesting seldon-core-loadtesting --repo https://storage.googleapis.com/seldon-charts     --set locust.host=http://seldon-model-example:8000     --set oauth.enabled=false     --set locust.hatchRate=1     --set locust.clients=1     --set loadtest.sendFeedback=0     --set locust.minWait=0     --set locust.maxWait=0     --set replicaCount=1')




# After a few mins you should see the deployment scaled to 5 replicas



# In[67]:


import json
import time


def getNumberPods():
    dp = get_ipython().getoutput('kubectl get deployment seldon-model-example-0-classifier -o json')
    dp = json.loads("".join(dp))
    return dp["status"]["replicas"]


scaled = False
for i in range(60):
    pods = getNumberPods()
    print(pods)
    if pods > 1:
        scaled = True
        break
    time.sleep(5)
assert scaled






# In[68]:


get_ipython().system('kubectl get deployment/seldon-model-example-0-classifier scaledobject/seldon-model-example-0-classifier')




# ## Remove Load


time.sleep(10)

# In[69]:


get_ipython().system('helm delete seldon-core-loadtesting')


time.sleep(5)


# After 5-10 mins you should see the deployment replica number decrease to 1

# ## Cleanup


time.sleep(10)

# In[71]:


get_ipython().system('kubectl delete -f model_with_keda_prom.yaml')


time.sleep(5)



time.sleep(10)

# In[72]:


get_ipython().system('helm delete seldon-monitoring -n seldon-monitoring')


time.sleep(5)



time.sleep(10)

# In[73]:


get_ipython().system('kubectl delete namespace seldon-monitoring')


time.sleep(5)




# In[ ]:
