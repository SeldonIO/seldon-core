# Install Seldon Core on EKS via AWS MarketPlace

 * Subscribe to Seldon Core on [AWS MarketPlace](https://aws.amazon.com/marketplace/seller-profile?id=cec67450-7a7e-43d5-8e5f-61e94e7c9e03&ref=dtl_B07KCNBCHV) and retrieve the log in command to authenticate your Docker client.

 (Note: below is for the 0.5.0 release and will differ for other releases).
  ```bash
  $(aws ecr get-login --no-include-email --region us-east-1 --registry-ids 403495124976)
  ```

 * [Create your EKS cluster and authenticate kubectl](https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html).
   Configure local Kubectl
  ```bash
  aws eks --region <CLUSTER_REGION> update-kubeconfig --name <CLUSTER-NAME>
  ```

 * Install [helm](https://docs.helm.sh/) on your cluster if it is not there already. **Note: Helm v3 is used for the following set up.**  

  Create a namespace for the Seldon system.

  ```bash
  kubectl create namespace seldon-system
  ```

 * Install Seldon Core for the release you subscribed to on Amazon MarketPlace:

For **Seldon 0.5.0**

 ```bash
  helm install seldon-core seldon-core-aws --repo https://storage.googleapis.com/seldon-aws-charts --version 0.5.0 --set usageMetrics.enabled=true --namespace seldon-system
 ```

## Next Steps

Follow the rest of the [install docs](../workflow/install.md).

