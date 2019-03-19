# Seldon Core Roadmap

The high level roadmap for Seldon Core. Feedback and additions very welcome.


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

  - ML Examples for various toolkits
  - Core components
    - Outlier Detection
    - Concept drift
    - Model Explanation 
  - E2E pipeline examples
    - Kubeflow/Kubeflow Pipelines
  - CI/CD 
  - GPU/TPU 
  - Deployment
    - Rolling, Canary, Red-Green

## Developer (required for 1.0)

  - Unit Test Coverage
  - E2E tests
  - CI pipelines (Prow)
  - Automated release process
    - Core, Python PyPI, AWS Marketplace, Google Marketplace, Kubeflow, helm charts, ksonnet
  - New Documentation site

## Specific Functionality

Please provide feedback and additions for helping us decide priorities.

 - Autoscaling
 - Kubeflow Pipelines integration
 - No service orchestrator for single model deployments
 - Wider range of image build integrations: e.g., Kaniko, img, GCB
 - NVIDIA Rapids, Dali integration
 - Julia, C++ wrappers
 - Shadow deployments
 - Edge deployments
 - Use of Apache Arrow for zero-copy, zero bandwidth RPC
 - Further Istio integration
 - Nginx-ingress
 - Openshift integration
 - Serverless/KNative
 