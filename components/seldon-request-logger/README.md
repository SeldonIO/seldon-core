# Example Request Logger

For use with request logging example (/examples/centralised-logging/request-logging/). The deployment yaml for this is there.

This example request logger is as general-purpose as possible. Eventually intended to read all types of CloudEvents emitted by the Seldon executor. These can be:

 - SeldonMessages or tensorflow format
 - Tabular data, string data or image data
 - Requests, responses to main predictor or to canary or shadow (all to be recorded in traceable way)
 - GRPC or json

Custom request loggers can be built for different types of transformations. Can be written in any language, just needs to handle HTTP POST requests containing cloud events and can log to chosen backend.

# Design

A SeldonDeployment or KFServing InferenceService is designed to output a dump of a request or response via http. Each includes headers indicating the source and the request id.

Different models and different endpoints are logged to different indexes. This ensures that models with different data formats won't conflict in document formats.

Different endpoints in different indexes also allows us to handle shadows. A shadow of a model will receive the same request but may have a different output. So we log both sets under different indexes.

Batch requests are split and indexed as separate documents with `-item-N` appended to the doc id.

See comments in log_helper.py for details of headers.

Large images are not recommended to be stored in elastic and instead a pointer should be stored. For this reason MAX_PAYLOAD_BYTES env var is available and defaults to only allow smaller images.

# Local Testing
First run elastic locally
```
docker pull docker.elastic.co/elasticsearch/elasticsearch-oss:7.6.0
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch-oss:7.6.0
```
Run the logger:
```
make run_local
```
And in another window run app/test.sh

The output of the logger will say where docs are indexed. The contents can be checked in postman by querying on the elastic host e.g.
```
localhost:9200/inference-log-seldon-unknown-namespace-strdata-unknown-endpoint/inferencerequest/7g
```

# Local Testing Against Seldon Executor

To try this out, run elastic locally
```
docker pull docker.elastic.co/elasticsearch/elasticsearch-oss:7.6.0
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch-oss:7.6.0
```

Run seldon-core/executor/samples/local/logger but don't start the dummy_logsink. Instead run the local request logger with:
```
make run_local
```
The log output from the request-logger will show the document id and index name, which is built from the seldon deployment name and namespace (if supplied).

View the document at `localhost:9200/<index>/inferencerequest/<doc_id>`

Example output is:

```
{
    "_index": "inference-log-seldon-default-seldon-single-model",
    "_type": "inferencerequest",
    "_id": "a8ea9850-7102-42c7-80d0-a6c26f1d8159",
    "_version": 4,
    "_seq_no": 3,
    "_primary_term": 1,
    "found": true,
    "_source": {
        "response": {
            "payload": {
                "meta": {},
                "data": {
                    "names": [
                        "proba"
                    ],
                    "ndarray": [
                        [
                            0.1951846770138402
                        ]
                    ]
                }
            },
            "dataType": "tabular",
            "elements": {
                "proba": [
                    0.1951846770138402
                ]
            },
            "ce-time": "2020-01-31T11:01:23.607035762Z",
            "ce-source": "http://localhost:8000/"
        },
        "ServingEngine": "Seldon",
        "Predictor": "example",
        "Namespace": "default",
        "Model-Id": "classifier",
        "request": {
            "payload": {
                "meta": {},
                "data": {
                    "ndarray": [
                        [
                            1.0,
                            2.0
                        ]
                    ]
                }
            },
            "dataType": "tabular",
            "elements": {},
            "ce-time": "2020-01-31T11:01:23.593571905Z",
            "ce-source": "http://localhost:8000/"
        }
    }
}
```


# On-going work

TODO: SOURCE IS ALWAYS http://localhost:8000/ WHEN COMING FROM EXECUTOR

TODO: HANDLE GRPC
TODO: THINK ABOUT SHADOW CASE - HOW TO ENSURE WE HAVE SOMETHING TO LINK DEFAULT AND SHADOW? MULTIPLE SHADOWS? https://github.com/SeldonIO/seldon-core/issues/1207
TODO: FEEDBACK - IF SENT WITH CUSTOM ID HEADER COULD SUPPORT A/B TESTS WITH RECORDED RESULTS
