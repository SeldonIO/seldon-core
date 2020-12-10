# Seldon Core Concepts

This page is a work in progress to provide concepts related to Seldon Core.
This work is ongoing and we welcome feedback.

## Machine Learning Deployment / Inference Graph

### Conceptual overview of machine learning deployments / inference graphs

A machine learning deployment (or inference graph) refers to a group of components in the Seldon ecosystem of a type associated with Seldon (Seldon Deployments). It represents a workflow, grouping the components of a machine learning system into a logical pipeline. The ML Deployment contains the configuration of the components and the definitions of the inputs and outputs of the system, and of each component.

## Component / Inference server

### Conceptual overview of components / inference servers

A component (or inference server) is one of:
-   Model
    
-   Router
    
-   Combiner
    
-   Transformer
    
-   Output Transformer

A program that performs one step in the workflow.

## Model

### Conceptual overview of models

A component within a machine learning deployment which holds the representation of learned information from the training data.

## Language Wrapper

### Conceptual overview of language wrappers

A language wrapper is a model which enables cross language and/or runtime interoperability with a particular programming language.

## Pre-packaged Inference Server

### Conceptual overview of pre-packaged inference servers

A pre-packaged inference server is one of:

-   SKLearn Server
    
-   XGBoost Server
    
-   Tensorflow Serving
    
-   MLflow Server
    
Servers which can be used to deploy a trained model.

Pre-packaged inference servers come built into Seldon Core, to allow users to go easily from artifact (i.e. serialised model) to ML deployment regardless of toolkit. Please see the following docs pages for users looking for instructions on how to create their own "pre-packaged" inference servers:
- [https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html](https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html)
- [https://docs.seldon.io/projects/seldon-core/en/latest/servers/custom.html](https://docs.seldon.io/projects/seldon-core/en/latest/servers/custom.html)

## Graph

### Conceptual overview of graphs

A graph represents machine learning components as nodes with edges representing the inputs and outputs of an operation being passed from one component to the next.

## Request

### Conceptual overview of requests

A request represents a single call to a model for a prediction. The request will be a payload which holds prediction data (often in the form of an array) passed over a particular protocol. It is expected to follow a particular format.

## Request Logging

### Conceptual overview of request logging

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
