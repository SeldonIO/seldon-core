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

See comments in log_helper.py for details of headers. The unmodified payload is logged under `request.payload` and the particular record under `request.instance`.

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
localhost:9200/inference-log-seldon-unknown-namespace-strdata-unknown-endpoint/_doc/7g
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

View the document at `localhost:9200/<index>/_doc/<doc_id>`

Here is a truncated version of an example output for an image use-case with outlier:

```
{
    "_index": "inference-log-inferenceservice-default-cifar10-default",
    "_type": "_doc",
    "_id": "9i",
    "_version": 3,
    "_seq_no": 2,
    "_primary_term": 1,
    "found": true,
    "_source": {
        "request": {
            "instance": [
                [
                    [
                        0.23137255012989044,
                        0.24313725531101227,
                        0.24705882370471954
                    ],
                    [
                        0.16862745583057404,
                        0.18039216101169586,
                        0.1764705926179886
                    ],
                    [
                        0.19607843458652496,
                        0.1882352977991104,
                        0.16862745583057404
                    ],
                    [
                        0.2666666805744171,
                        0.21176470816135406,
                        0.16470588743686676
                    ],
                    ...<TRUNCATED>...
                ]
            ],
            "dataType": "image",
            "payload": {
                "instances": [
                    [
                        [
                            [
                                0.23137255012989044,
                                0.24313725531101227,
                                0.24705882370471954
                            ],
                            [
                                0.16862745583057404,
                                0.18039216101169586,
                                0.1764705926179886
                            ],
                            [
                                0.19607843458652496,
                                0.1882352977991104,
                                0.16862745583057404
                            ],
                            [
                                0.2666666805744171,
                                0.21176470816135406,
                                0.16470588743686676
                            ],
                            ...<TRUNCATED>...
                        ]
                    ]
                ]
            }
        },
        "ServingEngine": "inferenceservice",
        "Ce-Inferenceservicename": "cifar10",
        "Ce-Endpoint": "default",
        "Ce-Namespace": "default",
        "@timestamp": "2020-04-09T08:15:20.923625+00:00",
        "RequestId": "9i",
        "response": {
            "instance": 2,
            "payload": {
                "predictions": [
                    2
                ]
            },
            "dataType": "tabular"
        },
        "outlier": {
            "data": {
                "feature_score": null,
                "instance_score": null,
                "is_outlier": 1
            },
            "meta": {
                "name": "OutlierVAE",
                "data_type": "image",
                "detector_type": "offline"
            }
        }
    }
}
```


# On-going work


TODO: HANDLE GRPC
TODO: ELEMENTS ARE CURRENTLY CREATED TO SPLIT FEATURES AND MAKE SEARCHABLE BY FEATURE VALUE. BUT ONLY FOR NON-BATCHED.
TODO: FEEDBACK - IF SENT WITH CUSTOM ID HEADER COULD SUPPORT A/B TESTS WITH RECORDED RESULTS
TODO: SOURCE IS ALWAYS http://localhost:8000/ WHEN COMING FROM EXECUTOR. PROB NOT LOGGER PROBLEM.