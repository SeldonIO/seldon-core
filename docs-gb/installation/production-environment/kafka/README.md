# Kafka Integration

[Kafka](https://kafka.apache.org/) is a component in the Seldon Core 2 ecosystem, that provides scalable, reliable, and flexible communication for machine learning deployments. It serves as a strong backbone for building complex inference pipelines, managing high-throughput asynchronous predictions, and seamlessly integrating with event-driven systems—key features needed for contemporary enterprise-grade ML platforms.

An inference request is a request sent to a machine learning model to make a prediction or inference based on input data. It is a core concept in deploying machine learning models in production, where models serve predictions to users or systems in real-time or batch mode.

To explore this feature of Seldon Core 2, you need to integrate with Kafka. Integrate Kafka through [managed cloud services](managed-kafka.md) or by deploying it [directly within a Kubernetes cluster](#self-hosted-kafka).

{% hint style="info" %}
**Note**: Kafka is an external component outside of the main Seldon stack. Therefore, it is the cluster administrator’s responsibility to administrate and manage the Kafka instance used by Seldon. For production installation it is highly recommended to use managed Kafka instance.
{% endhint %}

* [Securing Kafka](managed-kafka.md#securing-managed-kafka-services) provides more information about the encrytion and authentication.
* [Configuration examples](managed-kafka.md#example-configurations-for-managed-kafka-services) provides the steps to configure some of the managed Kafka services.

### Self-hosted Kafka

Seldon Enterprise Platform and Seldon Core 2 requires Kafka to implement data-centric inference Pipelines. To install Kafka for testing purposed in your Kubernetes cluster, use [Strimzi Operator](https://strimzi.io/docs/operators/latest/deploying). For more information, see [Self Hosted Kafka](/docs-gb/installation/learning-environment/self-hosted-kafka.md)
