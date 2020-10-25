#!/usr/bin/env ipython
import time
#!/usr/bin/env python
# coding: utf-8

# # Defining Disruption Budgets for Seldon Deployments

# ## Prerequisites
#  
# * A kubernetes cluster with kubectl configured
# * pygmentize

# ## Setup Seldon Core
# 
# Use the setup notebook to [Setup Cluster](../../../notebooks/seldon_core_setup.ipynb#Setup-Cluster) with [Ambassador Ingress](../../../notebooks/seldon_core_setup.ipynb#Ambassador) and [Install Seldon Core](../../seldon_core_setup.ipynb#Install-Seldon-Core). Instructions [also online](./seldon_core_setup.html).



# In[ ]:


get_ipython().system('kubectl create namespace seldon')






# In[ ]:


get_ipython().system('kubectl config set-context $(kubectl config current-context) --namespace=seldon')




# ## Create model with Pod Disruption Budget
# 
# To create a model with a Pod Disruption Budget, it is first important to understand how you would like your application to respond to [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).  Depending on the type of disruption budgeting your application needs, you will either define either of the following:
# 
# * `minAvailable` which is a description of the number of pods from that set that must still be available after the eviction, even in the absence of the evicted pod. `minAvailable` can be either an absolute number or a percentage.
# * `maxUnavailable` which is a description of the number of pods from that set that can be unavailable after the eviction. It can be either an absolute number or a percentage.
# 
# The full SeldonDeployment spec is shown below.



# In[ ]:


get_ipython().system('pygmentize model_with_pdb.yaml')






# In[ ]:


get_ipython().system('kubectl apply -f model_with_pdb.yaml')





time.sleep(10)

# In[ ]:


get_ipython().system("kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')")


time.sleep(5)


# ## Validate Disruption Budget Configuration



# In[ ]:


import json

def getPdbConfig():
    dp = get_ipython().getoutput('kubectl get pdb seldon-model-example-0-classifier -o json')
    dp=json.loads("".join(dp))
    return dp["spec"]["maxUnavailable"]
    
assert getPdbConfig() == 2






# In[ ]:


get_ipython().system('kubectl get pods,deployments,pdb')




# ## Update Disruption Budget and Validate Change
# 
# Next, we'll update the maximum number of unavailable pods and check that the PDB is properly updated to match.



# In[ ]:


get_ipython().system('pygmentize model_with_patched_pdb.yaml')






# In[ ]:


get_ipython().system('kubectl apply -f model_with_patched_pdb.yaml')





time.sleep(10)

# In[ ]:


get_ipython().system("kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')")


time.sleep(5)




# In[ ]:


assert getPdbConfig() == 1




# ## Clean Up



# In[ ]:


get_ipython().system('kubectl get pods,deployments,pdb')





time.sleep(10)

# In[ ]:


get_ipython().system('kubectl delete -f model_with_patched_pdb.yaml')


time.sleep(5)




# In[ ]:






