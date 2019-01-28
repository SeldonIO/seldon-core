# Seldon Core Roadmap

The high level roadmap for Seldon Core. Feedback and additions very welcome.

&#9745; : Done or In Progress

## Core (required for 1.0)

  - SeldonDeployment K8S CRD stability
  - API stability
     - External API (predict, feedback endpoints)
     - Internal microservice API (predict, transform, route etc.)
  - Wrapper stability for core wrappers
     - Python reference wrapper
  - Metrics and logging stability
  - Benchmarks for ongoing optimisation 

## Reference Examples (desired for 1.0)

  - &#9745; ML Examples for various toolkits
  - Core components
    - Outlier Detection
    - Concept drift
    - Model Explanation 
  - E2E pipeline examples
    - Kubeflow/Kubeflow Pipelines
  - &#9745; CI/CD 
  - GPU/TPU 
  - &#9745; Deployment
    - Rolling, Canary, Red-Green

## Developer (required for 1.0)

  - &#9745; Unit Test Coverage
  - &#9745; E2E tests
  - &#9745; CI pipelines (Prow)
  - Automated release process
    - Core, Python PyPI, AWS Marketplace, Google Marketplace, Kubeflow, helm charts, ksonnet
  - New Documentation site

## Specific Functionality

Please provide feedback and additions for helping us decide priorities.

 - &#9745; Autoscaling
 - Kubeflow Pipelines integration
 - No service orchestrator for single model deployments
 - Wider range of image build integrations: e.g., Kaniko, img, GCB
 - NVIDIA Rapids, Dali integration
 - Julia, C++ wrappers
 - &#9745; Shadow deployments
 - Edge deployments
 - Use of Apache Arrow for zero-copy, zero bandwidth RPC
 - Further Istio integration
 - Nginx-ingress
 - Openshift integration
 - Serverless/KNative
 