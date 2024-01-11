#!/usr/bin/env ipython
import time
#!/usr/bin/env python
# coding: utf-8

# # Stateful Model Feedback Metrics Server
# In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server.
#
# Dependencies
# * Seldon Core installed
# * Ingress provider (Istio or Ambassador)
#
# An easy way is to run `examples/centralized-logging/full-kind-setup.sh` and then:
# ```bash
#     helm delete seldon-core-loadtesting
#     helm delete seldon-single-model
# ```
#
# Then port-forward to that ingress on localhost:8003 in a separate terminal either with:
#
# Ambassador:
#
#     kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080
#
# Istio:
#
#     kubectl port-forward -n istio-system svc/istio-ingressgateway 8003:80
#
#
#



# In[1]:


get_ipython().system('kubectl create namespace seldon || echo "namespace already created"')






# In[2]:


get_ipython().system('kubectl config set-context $(kubectl config current-context) --namespace=seldon')






# In[3]:


get_ipython().system('mkdir -p config')




# ### Create a simple model
# We create a multiclass classification model - iris classifier.
#
# The iris classifier takes an input array, and returns the prediction of the 4 classes.
#
# The prediction can be done as numeric or as a probability array.



# In[4]:


get_ipython().run_cell_magic('bash', '', 'kubectl apply -f - << END\napiVersion: machinelearning.seldon.io/v1\nkind: SeldonDeployment\nmetadata:\n  name: multiclass-model\nspec:\n  predictors:\n  - graph:\n      children: []\n      implementation: SKLEARN_SERVER\n      modelUri: gs://seldon-models/v1.18.0/sklearn/iris\n      name: classifier\n      logger:\n        url: http://seldon-multiclass-model-metrics.seldon.svc.cluster.local:80/\n        mode: all\n    name: default\n    replicas: 1\nEND')





time.sleep(10)

# In[5]:


get_ipython().system("kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=multiclass-model -o jsonpath='{.items[0].metadata.name}')")


time.sleep(5)


# #### Send test request



# In[8]:


res = get_ipython().getoutput('curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/predictions"         -H "Content-Type: application/json" -d \'{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}\'')
print(res)
import json
j=json.loads(res[-1])
assert(len(j["data"]["ndarray"][0])==3)




# ### Metrics Server
# You can create a kubernetes deployment of the metrics server with this:



# In[9]:


get_ipython().run_cell_magic('writefile', 'config/multiclass-deployment.yaml', 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: seldon-multiclass-model-metrics\n  namespace: seldon\n  labels:\n    app: seldon-multiclass-model-metrics\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: seldon-multiclass-model-metrics\n  template:\n    metadata:\n      labels:\n        app: seldon-multiclass-model-metrics\n    spec:\n      securityContext:\n          runAsUser: 8888\n      containers:\n      - name: user-container\n        image: seldonio/alibi-detect-server:1.18.0\n        imagePullPolicy: IfNotPresent\n        args:\n        - --model_name\n        - multiclassserver\n        - --http_port\n        - \'8080\'\n        - --protocol\n        - seldonfeedback.http\n        - --storage_uri\n        - "adserver.cm_models.multiclass_one_hot.MulticlassOneHot"\n        - --reply_url\n        - http://message-dumper.default        \n        - --event_type\n        - io.seldon.serving.feedback.metrics\n        - --event_source\n        - io.seldon.serving.feedback\n        - MetricsServer\n        env:\n        - name: "SELDON_DEPLOYMENT_ID"\n          value: "multiclass-model"\n        - name: "PREDICTIVE_UNIT_ID"\n          value: "classifier"\n        - name: "PREDICTIVE_UNIT_IMAGE"\n          value: "seldonio/alibi-detect-server:1.18.0"\n        - name: "PREDICTOR_ID"\n          value: "default"\n---\napiVersion: v1\nkind: Service\nmetadata:\n  name: seldon-multiclass-model-metrics\n  namespace: seldon\n  labels:\n    app: seldon-multiclass-model-metrics\nspec:\n  selector:\n    app: seldon-multiclass-model-metrics\n  ports:\n    - protocol: TCP\n      port: 80\n      targetPort: 8080')






# In[10]:


get_ipython().system('kubectl apply -n seldon -f config/multiclass-deployment.yaml')





time.sleep(10)

# In[11]:


get_ipython().system('kubectl rollout status deploy/seldon-multiclass-model-metrics')


time.sleep(5)




# In[12]:


import time

time.sleep(20)




# ### Send feedback



# In[13]:


res = get_ipython().getoutput('curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/feedback"         -H "Content-Type: application/json"         -d \'{"response": {"data": {"ndarray": [[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]}}, "truth":{"data": {"ndarray": [[0,0,1]]}}}\'')
print(res)
import json
j=json.loads(res[-1])
assert("data" in j)






# In[14]:


import time

time.sleep(3)




# ### Check that metrics are recorded



# In[15]:


res = get_ipython().getoutput('kubectl logs $(kubectl get pods -l app=seldon-multiclass-model-metrics                     -n seldon -o jsonpath=\'{.items[0].metadata.name}\') | grep "PROCESSING Feedback Event"')
print(res)
assert(len(res)>0)




# ### Cleanup


time.sleep(10)

# In[19]:


get_ipython().system('kubectl delete -n seldon -f config/multiclass-deployment.yaml')


time.sleep(5)



time.sleep(10)

# In[20]:


get_ipython().system('kubectl delete sdep multiclass-model')


time.sleep(5)
