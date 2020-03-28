#!/usr/bin/env ipython
import time
#!/usr/bin/env python
# coding: utf-8

# # Scikit-Learn IRIS Model
# 
#  * Wrap a scikit-learn python model for use as a prediction microservice in seldon-core
#    * Run locally on Docker to test
#    * Deploy on seldon-core running on a kubernetes cluster
#  
# ## Dependencies
# 
#  * [S2I](https://github.com/openshift/source-to-image)
# 
# ```bash
# pip install sklearn
# pip install seldon-core
# ```
# 
# ## Train locally
#  



# In[ ]:


import numpy as np
import os
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline
from sklearn.externals import joblib
from sklearn import datasets

def main():
    clf = LogisticRegression()
    p = Pipeline([('clf', clf)])
    print('Training model...')
    p.fit(X, y)
    print('Model trained!')

    filename_p = 'IrisClassifier.sav'
    print('Saving model in %s' % filename_p)
    joblib.dump(p, filename_p)
    print('Model saved!')
    
if __name__ == "__main__":
    print('Loading iris data set...')
    iris = datasets.load_iris()
    X, y = iris.data, iris.target
    print('Dataset loaded!')
    main()




# Wrap model using s2i

# ## REST test



# In[ ]:


get_ipython().system('s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.18 seldonio/sklearn-iris:0.1')






# In[ ]:


get_ipython().system('docker run --name "iris_predictor" -d --rm -p 5000:5000 seldonio/sklearn-iris:0.1')


time.sleep(3)


# Send some random features that conform to the contract



# In[ ]:


get_ipython().system('curl  -s http://localhost:5000/predict -H "Content-Type: application/json" -d \'{"data":{"ndarray":[[5.964,4.006,2.081,1.031]]}}\'')






# In[ ]:


get_ipython().system('docker rm iris_predictor --force')




# ## Setup Seldon Core
# 
# Use the setup notebook to [Setup Cluster](seldon_core_setup.ipynb) to setup Seldon Core with an ingress - either Ambassador or Istio.
# 
# Then port-forward to that ingress on localhost:8003 in a separate terminal either with:
# 
#  * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
#  * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:80`



# In[ ]:


get_ipython().system('kubectl create namespace seldon')






# In[ ]:


get_ipython().system('kubectl config set-context $(kubectl config current-context) --namespace=seldon')






# In[ ]:


get_ipython().system('kubectl create -f sklearn_iris_deployment.yaml')





time.sleep(10)

# In[ ]:


get_ipython().system('kubectl rollout status deploy/seldon-deployment-example-sklearn-iris-predictor-0')


time.sleep(2)




# In[ ]:


res = get_ipython().getoutput('curl  -s http://localhost:8003/seldon/seldon/seldon-deployment-example/api/v0.1/predictions -H "Content-Type: application/json" -d \'{"data":{"ndarray":[[5.964,4.006,2.081,1.031]]}}\'')






# In[ ]:


res






# In[ ]:


print(res)
import json
j=json.loads(res[0])
assert(j["data"]["ndarray"][0][0]>0.0)





time.sleep(10)

# In[ ]:


get_ipython().system('kubectl delete -f sklearn_iris_deployment.yaml')


time.sleep(2)

