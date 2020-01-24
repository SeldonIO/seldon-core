# Example Request Logger

This example request logger is as general-purpose as possible. Eventually intended to read all types of CloudEvents emitted by the Seldon executor. These can be:

 - GRPC or json
 - SeldonMessages or tensorflow format
 - Tabular data, string data or image data
 - Requests, responses to main predictor or to canary or shadow (all to be aggregated)

Custom request loggers can be built for different types of transformations. Can be written in any language, just needs to handle HTTP POST requests and log to stdout for fluentd or could go direct to chosen backend.

To try this out, run elastic locally
```
docker pull docker.elastic.co/elasticsearch/elasticsearch:7.5.2
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.5.2
```

NEED TO UPDATE THE EXAMPLES IN test.sh

USE seldon-core/executor/samples/local/logger

ENV VARS NEEDED

NEED TO SET A PROPER ID