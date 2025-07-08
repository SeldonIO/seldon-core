# Example Seldon Core Deployments using Helm
<img src="images/deploy-graph.png" alt="predictor with canary" title="ml graph"/>

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.


## Serve Single Model


```python
!helm install mymodel ../helm-charts/seldon-single-model --set 'model.image=seldonio/mock_classifier:1.5.0-dev'
```

    NAME: mymodel
    LAST DEPLOYED: Mon Nov  2 11:18:38 2020
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template mymodel ../helm-charts/seldon-single-model --set 'model.image=seldonio/mock_classifier:1.5.0-dev' | pygmentize -l json
```

    [04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m
    [04m[31;01m#[39;49;00m [04m[31;01mS[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mu[39;49;00m[04m[31;01mr[39;49;00m[04m[31;01mc[39;49;00m[04m[31;01me[39;49;00m[04m[31;01m:[39;49;00m [04m[31;01ms[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01md[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mi[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01mg[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01me[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01md[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01mp[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01md[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01md[39;49;00m[04m[31;01me[39;49;00m[04m[31;01mp[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01my[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01me[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01m.[39;49;00m[04m[31;01mj[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m
    {
      [34;01m"kind"[39;49;00m: [33m"SeldonDeployment"[39;49;00m,
      [34;01m"apiVersion"[39;49;00m: [33m"machinelearning.seldon.io/v1"[39;49;00m,
      [34;01m"metadata"[39;49;00m: {
        [34;01m"name"[39;49;00m: [33m"mymodel"[39;49;00m,
        [34;01m"namespace"[39;49;00m: [33m"seldon"[39;49;00m,
        [34;01m"labels"[39;49;00m: {}
      },
      [34;01m"spec"[39;49;00m: {
          [34;01m"name"[39;49;00m: [33m"mymodel"[39;49;00m,
          [34;01m"protocol"[39;49;00m: [33m"seldon"[39;49;00m,
        [34;01m"annotations"[39;49;00m: {},
        [34;01m"predictors"[39;49;00m: [
          {
            [34;01m"name"[39;49;00m: [33m"default"[39;49;00m,
            [34;01m"graph"[39;49;00m: {
              [34;01m"name"[39;49;00m: [33m"model"[39;49;00m,
              [34;01m"type"[39;49;00m: [33m"MODEL"[39;49;00m,
            },
            [34;01m"componentSpecs"[39;49;00m: [
              {
                [34;01m"spec"[39;49;00m: {
                  [34;01m"containers"[39;49;00m: [
                    {
                      [34;01m"name"[39;49;00m: [33m"model"[39;49;00m,
                      [34;01m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.5.0-dev"[39;49;00m,
                      [34;01m"env"[39;49;00m: [
                          {
                            [34;01m"name"[39;49;00m: [33m"LOG_LEVEL"[39;49;00m,
                            [34;01m"value"[39;49;00m: [33m"INFO"[39;49;00m
                          },
                        ],
                      [34;01m"resources"[39;49;00m: {[34;01m"requests"[39;49;00m:{[34;01m"memory"[39;49;00m:[33m"1Mi"[39;49;00m}},
                    }
                  ]
                },
              }
            ],
            [34;01m"replicas"[39;49;00m: [34m1[39;49;00m
          }
        ]
      }
    }



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=mymodel -o jsonpath='{.items[0].metadata.name}')
```

    deployment "mymodel-default-0-model" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="mymodel",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request


```python
r = sc.predict(transport="rest")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.0406846384836026
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05335370865277927]}}, 'meta': {}}


#### GRPC Request


```python
r = sc.predict(transport="grpc")
print(r)
```

    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.3321428950191112]}}}
    Response:
    {'meta': {}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.07014111011256721]}}}



```python
!helm delete mymodel
```

    release "mymodel" uninstalled


## Serve REST AB Test


```python
!helm install myabtest ../helm-charts/seldon-abtest
```

    NAME: myabtest
    LAST DEPLOYED: Mon Nov  2 11:19:50 2020
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template ../helm-charts/seldon-abtest | pygmentize -l json
```

    [04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m
    [04m[31;01m#[39;49;00m [04m[31;01mS[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mu[39;49;00m[04m[31;01mr[39;49;00m[04m[31;01mc[39;49;00m[04m[31;01me[39;49;00m[04m[31;01m:[39;49;00m [04m[31;01ms[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01md[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mb[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01mp[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mb[39;49;00m[04m[31;01m_[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01m_[39;49;00m[34m2[39;49;00m[04m[31;01mp[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01md[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01m.[39;49;00m[04m[31;01mj[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m
    {
        [34;01m"apiVersion"[39;49;00m: [33m"machinelearning.seldon.io/v1alpha2"[39;49;00m,
        [34;01m"kind"[39;49;00m: [33m"SeldonDeployment"[39;49;00m,
        [34;01m"metadata"[39;49;00m: {
    	[34;01m"labels"[39;49;00m: {
    	    [34;01m"app"[39;49;00m: [33m"seldon"[39;49;00m
    	},
    	[34;01m"name"[39;49;00m: [33m"RELEASE-NAME"[39;49;00m
        },
        [34;01m"spec"[39;49;00m: {
    	[34;01m"name"[39;49;00m: [33m"RELEASE-NAME"[39;49;00m,
    	[34;01m"predictors"[39;49;00m: [
    	    {
    		[34;01m"name"[39;49;00m: [33m"default"[39;49;00m,
    		[34;01m"replicas"[39;49;00m: [34m1[39;49;00m,
    		[34;01m"componentSpecs"[39;49;00m: [{
    		    [34;01m"spec"[39;49;00m: {
    			[34;01m"containers"[39;49;00m: [
    			    {
                                    [34;01m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.5.0-dev"[39;49;00m,
    				[34;01m"imagePullPolicy"[39;49;00m: [33m"IfNotPresent"[39;49;00m,
    				[34;01m"name"[39;49;00m: [33m"classifier-1"[39;49;00m,
    				[34;01m"resources"[39;49;00m: {
    				    [34;01m"requests"[39;49;00m: {
    					[34;01m"memory"[39;49;00m: [33m"1Mi"[39;49;00m
    				    }
    				}
    			    }],
    			[34;01m"terminationGracePeriodSeconds"[39;49;00m: [34m20[39;49;00m
    		    }},
    	        {
    		    [34;01m"metadata"[39;49;00m:{
    			[34;01m"labels"[39;49;00m:{
    			    [34;01m"version"[39;49;00m:[33m"v2"[39;49;00m
    			}
    		    },    
    			[34;01m"spec"[39;49;00m:{
    			    [34;01m"containers"[39;49;00m:[
    				{
                                    [34;01m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.5.0-dev"[39;49;00m,
    				[34;01m"imagePullPolicy"[39;49;00m: [33m"IfNotPresent"[39;49;00m,
    				[34;01m"name"[39;49;00m: [33m"classifier-2"[39;49;00m,
    				[34;01m"resources"[39;49;00m: {
    				    [34;01m"requests"[39;49;00m: {
    					[34;01m"memory"[39;49;00m: [33m"1Mi"[39;49;00m
    				    }
    				}
    			    }
    			],
    			[34;01m"terminationGracePeriodSeconds"[39;49;00m: [34m20[39;49;00m
    				   }
    				   }],
    		[34;01m"graph"[39;49;00m: {
    		    [34;01m"name"[39;49;00m: [33m"RELEASE-NAME"[39;49;00m,
    		    [34;01m"implementation"[39;49;00m:[33m"RANDOM_ABTEST"[39;49;00m,
    		    [34;01m"parameters"[39;49;00m: [
    			{
    			    [34;01m"name"[39;49;00m:[33m"ratioA"[39;49;00m,
    			    [34;01m"value"[39;49;00m:[33m"0.5"[39;49;00m,
    			    [34;01m"type"[39;49;00m:[33m"FLOAT"[39;49;00m
    			}
    		    ],
    		    [34;01m"children"[39;49;00m: [
    			{
    			    [34;01m"name"[39;49;00m: [33m"classifier-1"[39;49;00m,
    			    [34;01m"type"[39;49;00m:[33m"MODEL"[39;49;00m,
    			    [34;01m"children"[39;49;00m:[]
    			},
    			{
    			    [34;01m"name"[39;49;00m: [33m"classifier-2"[39;49;00m,
    			    [34;01m"type"[39;49;00m:[33m"MODEL"[39;49;00m,
    			    [34;01m"children"[39;49;00m:[]
    			}   
    		    ]
    		}
    	    }
    	]
        }
    }



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=myabtest -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=myabtest -o jsonpath='{.items[1].metadata.name}')
```

    Waiting for deployment "myabtest-default-0-classifier-1" rollout to finish: 0 of 1 updated replicas are available...
    deployment "myabtest-default-0-classifier-1" successfully rolled out
    deployment "myabtest-default-1-classifier-2" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="myabtest",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request


```python
r = sc.predict(transport="rest")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.8562060281778412
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.11299965170860979]}}, 'meta': {}}


#### gRPC Request


```python
r = sc.predict(transport="grpc")
print(r)
```

    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.45187622094165814]}}}
    Response:
    {'meta': {}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.07836365822139986]}}}



```python
!helm delete myabtest
```

    release "myabtest" uninstalled


## Serve REST Multi-Armed Bandit


```python
!helm install mymab ../helm-charts/seldon-mab
```

    NAME: mymab
    LAST DEPLOYED: Mon Nov  2 11:22:19 2020
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template ../helm-charts/seldon-mab | pygmentize -l json
```

    [04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01m-[39;49;00m
    [04m[31;01m#[39;49;00m [04m[31;01mS[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mu[39;49;00m[04m[31;01mr[39;49;00m[04m[31;01mc[39;49;00m[04m[31;01me[39;49;00m[04m[31;01m:[39;49;00m [04m[31;01ms[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01md[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m[04m[31;01m-[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mb[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01mp[39;49;00m[04m[31;01ml[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mt[39;49;00m[04m[31;01me[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01m/[39;49;00m[04m[31;01mm[39;49;00m[04m[31;01ma[39;49;00m[04m[31;01mb[39;49;00m[04m[31;01m.[39;49;00m[04m[31;01mj[39;49;00m[04m[31;01ms[39;49;00m[04m[31;01mo[39;49;00m[04m[31;01mn[39;49;00m
    {
        [34;01m"apiVersion"[39;49;00m: [33m"machinelearning.seldon.io/v1alpha2"[39;49;00m,
        [34;01m"kind"[39;49;00m: [33m"SeldonDeployment"[39;49;00m,
        [34;01m"metadata"[39;49;00m: {
    		[34;01m"labels"[39;49;00m: {[34;01m"app"[39;49;00m:[33m"seldon"[39;49;00m},
    		[34;01m"name"[39;49;00m: [33m"RELEASE-NAME"[39;49;00m
        },
        [34;01m"spec"[39;49;00m: {
    	[34;01m"name"[39;49;00m: [33m"RELEASE-NAME"[39;49;00m,
    	[34;01m"predictors"[39;49;00m: [
    	    {
    		[34;01m"name"[39;49;00m: [33m"default"[39;49;00m,
    		[34;01m"replicas"[39;49;00m: [34m1[39;49;00m,
    		[34;01m"componentSpecs"[39;49;00m: [{
    		    [34;01m"spec"[39;49;00m: {
    			[34;01m"containers"[39;49;00m: [
    			    {
                                    [34;01m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.5.0-dev"[39;49;00m,				
    				[34;01m"imagePullPolicy"[39;49;00m: [33m"IfNotPresent"[39;49;00m,
    				[34;01m"name"[39;49;00m: [33m"classifier-1"[39;49;00m,
    				[34;01m"resources"[39;49;00m: {
    				    [34;01m"requests"[39;49;00m: {
    					[34;01m"memory"[39;49;00m: [33m"1Mi"[39;49;00m
    				    }
    				}
    			    }],
    			[34;01m"terminationGracePeriodSeconds"[39;49;00m: [34m20[39;49;00m
    		    }},
    	        {
    			[34;01m"spec"[39;49;00m:{
    			    [34;01m"containers"[39;49;00m:[
    				{
                                    [34;01m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.5.0-dev"[39;49;00m,								    
    				[34;01m"imagePullPolicy"[39;49;00m: [33m"IfNotPresent"[39;49;00m,
    				[34;01m"name"[39;49;00m: [33m"classifier-2"[39;49;00m,
    				[34;01m"resources"[39;49;00m: {
    				    [34;01m"requests"[39;49;00m: {
    					[34;01m"memory"[39;49;00m: [33m"1Mi"[39;49;00m
    				    }
    				}
    			    }
    			],
    			[34;01m"terminationGracePeriodSeconds"[39;49;00m: [34m20[39;49;00m
    			}
    		},
    	        {
    		    [34;01m"spec"[39;49;00m:{
    			[34;01m"containers"[39;49;00m: [{
                                [34;01m"image"[39;49;00m: [33m"seldonio/mab_epsilon_greedy:1.5.0-dev"[39;49;00m,								    			    
    			    [34;01m"name"[39;49;00m: [33m"eg-router"[39;49;00m
    			}],
    			[34;01m"terminationGracePeriodSeconds"[39;49;00m: [34m20[39;49;00m
    		    }}
    	        ],
    		[34;01m"graph"[39;49;00m: {
    		    [34;01m"name"[39;49;00m: [33m"eg-router"[39;49;00m,
    		    [34;01m"type"[39;49;00m:[33m"ROUTER"[39;49;00m,
    		    [34;01m"parameters"[39;49;00m: [
    			{
    			    [34;01m"name"[39;49;00m: [33m"n_branches"[39;49;00m,
    			    [34;01m"value"[39;49;00m: [33m"2"[39;49;00m,
    			    [34;01m"type"[39;49;00m: [33m"INT"[39;49;00m
    			},
    			{
    			    [34;01m"name"[39;49;00m: [33m"epsilon"[39;49;00m,
    			    [34;01m"value"[39;49;00m: [33m"0.2"[39;49;00m,
    			    [34;01m"type"[39;49;00m: [33m"FLOAT"[39;49;00m
    			},
    			{
    			    [34;01m"name"[39;49;00m: [33m"verbose"[39;49;00m,
    			    [34;01m"value"[39;49;00m: [33m"1"[39;49;00m,
    			    [34;01m"type"[39;49;00m: [33m"BOOL"[39;49;00m
    			}
    		    ],
    		    [34;01m"children"[39;49;00m: [
    			{
    			    [34;01m"name"[39;49;00m: [33m"classifier-1"[39;49;00m,
    			    [34;01m"type"[39;49;00m:[33m"MODEL"[39;49;00m,
    			    [34;01m"children"[39;49;00m:[]
    			},
    			{
    			    [34;01m"name"[39;49;00m: [33m"classifier-2"[39;49;00m,
    			    [34;01m"type"[39;49;00m:[33m"MODEL"[39;49;00m,
    			    [34;01m"children"[39;49;00m:[]
    			}   
    		    ]
    		},
    		[34;01m"svcOrchSpec"[39;49;00m: {
    		[34;01m"resources"[39;49;00m: {[34;01m"requests"[39;49;00m:{[34;01m"cpu"[39;49;00m:[33m"0.1"[39;49;00m}},
    [34;01m"env"[39;49;00m: [
    {
    [34;01m"name"[39;49;00m: [33m"SELDON_LOG_MESSAGES_EXTERNALLY"[39;49;00m,
    [34;01m"value"[39;49;00m: [33m"false"[39;49;00m
    },
    {
    [34;01m"name"[39;49;00m: [33m"SELDON_LOG_MESSAGE_TYPE"[39;49;00m,
    [34;01m"value"[39;49;00m: [33m"seldon.message.pair"[39;49;00m
    },
    {
    [34;01m"name"[39;49;00m: [33m"SELDON_LOG_REQUESTS"[39;49;00m,
    [34;01m"value"[39;49;00m: [33m"false"[39;49;00m
    },
    {
    [34;01m"name"[39;49;00m: [33m"SELDON_LOG_RESPONSES"[39;49;00m,
    [34;01m"value"[39;49;00m: [33m"false"[39;49;00m
    },
    ]
    },
    		[34;01m"labels"[39;49;00m: {[34;01m"fluentd"[39;49;00m:[33m"true"[39;49;00m,[34;01m"version"[39;49;00m:[33m"1.5.0-dev"[39;49;00m}
    	    }
    	]
        }
    }



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=mymab -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=mymab -o jsonpath='{.items[1].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=mymab -o jsonpath='{.items[2].metadata.name}')
```

    Waiting for deployment "mymab-default-0-classifier-1" rollout to finish: 0 of 1 updated replicas are available...
    deployment "mymab-default-0-classifier-1" successfully rolled out
    deployment "mymab-default-1-classifier-2" successfully rolled out
    deployment "mymab-default-2-eg-router" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="mymab",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request


```python
r = sc.predict(transport="rest")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.1000299187972008
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05643175042558145]}}, 'meta': {}}


#### gRPC Request


```python
r = sc.predict(transport="grpc")
print(r)
```

    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.23579893772394123]}}}
    Response:
    {'meta': {}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0641117916962909]}}}



```python
!helm delete mymab
```

    release "mymab" uninstalled



```python

```
