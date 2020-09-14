# Seldon Core Concepts

This page is a work in progress to provide concepts related to Seldon Core.

This work is ongoing and we welcome feedback

## Workflow
#### Conceptual overview of workflows
A workflow groups the components of a machine learning system into a logical pipeline. Sets of related components are grouped into graphs. The workflow contains the configuration of the components and the definitions of the inputs and outputs of the system, and of each component.

## Component
#### Conceptual overview of components
A component is one of:
-   Model
    
-   Router
    
-   Combiner
    
-   Transformer
    
-   Output Transformer

A program that performs one step in the workflow.

## Machine Learning Deployment
#### Conceptual overview of machine learning deployments
A machine learning deployment refers to a predictive model in the Seldon ecosystem of a type associated with Seldon (Seldon Deployments) or Kubeflow (Inference Services).

## Model
#### Conceptual overview of models
A component within a machine learning deployment which holds the representation of learned information from the training data.

## Language Wrapper
#### Conceptual overview of language wrappers
Enables cross language and/or runtime interoperability of a model with a particular programming language.

## Pre-packaged Inference Server
#### Conceptual overview of pre-packaged inference servers
A pre-packaged inference server is one of:

-   SKLearn Server
    
-   XGBoost Server
    
-   Tensorflow Serving
    
-   MLflow Server
    
Servers which can be used to deploy a trained model.

## Graph
#### Conceptual overview of graphs
A graph represents machine learning components as nodes with edges representing the inputs and outputs of an operation being passed from one component to the next.

## Request
#### Conceptual overview of requests
A request represents a single call to a model for a prediction. The request will be a payload which holds prediction data (often in the form of an array) passed over a particular protocol.

## Request Logging
#### Conceptual overview of request logging
Request logging is a feature of Seldon to be able to track the requests that have been submitted to a model. In the default setup requests are logged and stored in Elasticsearch.

## Useful Links

### Kubeflow pipelines concepts

[https://www.kubeflow.org/docs/pipelines/overview/concepts/](https://www.kubeflow.org/docs/pipelines/overview/concepts/)

  

### Google machine learning glossary

[https://developers.google.com/machine-learning/glossary](https://developers.google.com/machine-learning/glossary)

  

### Kubernetes standardized glossary

[https://kubernetes.io/docs/reference/glossary/?fundamental=true](https://kubernetes.io/docs/reference/glossary/?fundamental=true)

  

### Helm glossary

[https://helm.sh/docs/glossary/](https://helm.sh/docs/glossary/)

  

### Jaeger terminology

[https://www.jaegertracing.io/docs/1.18/architecture/#terminology](https://www.jaegertracing.io/docs/1.18/architecture/#terminology)

  

### Elastic search terminology

[https://www.elastic.co/guide/en/elastic-stack-glossary/current/terms.html](https://www.elastic.co/guide/en/elastic-stack-glossary/current/terms.html)
