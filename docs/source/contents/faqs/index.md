# FAQs

## Can Seldon Core V2 be used with Seldon Core V1

The two projects are able to be run side by side. Existing users of Seldon Core can update to deploy their models using Seldon Core V2 as needed to take advantage of the new functionality. Both will be supported.

## Should I choose V1 APIs or V2 APIs

This depends on your use case. V2 APIs are presently alpha so not yet at GA so might contain breaking changes in future releases.

 Use V1 for:

  * Tight integration to Seldon V1 protocol
  * Tensorflow Server requirements
  * Need managed istio integration

 Use V2 for:

  * Multi-model serving
  * More expressive DAG inference pipelines
  * Data-centric (Kafka)
  * Service mesh agnostic
  * Simpler single model usage
  * V2 Protocol

## Can I do payload logging in Seldon Core v2?

By default, the input and output of every step in a pipeline (as well as the pipeline itself) is logged in Kafka.
From there it's up to you what to do with the data.
You could use something like [kafka connect](https://docs.confluent.io/platform/current/connect/index.html) to stream the logs to a datastore of your choice. 

Note that direct inference to a model outside of a pipeline will not automatically be logged.

