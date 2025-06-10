# FAQs

## Can I do payload logging in Seldon Core 2?

By default, the input and output of every step in a pipeline (as well as the pipeline itself) is logged in Kafka.
From there it's up to you what to do with the data.
You could use something like [Kafka Connect](https://docs.confluent.io/platform/current/connect/index.html) to stream the logs to a datastore of your choice. 

Note that there is no automatic request logging for models being accessed directly over REST or gRPC.
Requests need to be sent via pipelines to be recorded in Kafka.

