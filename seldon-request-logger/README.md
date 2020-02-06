# Example Request Logger

For use with request logging example (/examples/centralised-logging/request-logging/). The deployment yaml for this is there.

This example request logger is as general-purpose as possible. Eventually intended to read all types of CloudEvents emitted by the Seldon executor. These can be:

 - GRPC or json
 - SeldonMessages or tensorflow format
 - Tabular data, string data or image data
 - Requests, responses to main predictor or to canary or shadow (all to be aggregated)

Custom request loggers can be built for different types of transformations. Can be written in any language, just needs to handle HTTP POST requests and log to stdout for fluentd or could go direct to chosen backend.

# Local Testing

To try this out, run elastic locally
```
docker pull docker.elastic.co/elasticsearch/elasticsearch:7.5.2
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.5.2
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

TODO: UPDATE CENTRALISED LOGGING EXAMPLE - INC PUBLISHING IMAGE
TODO: DOCUMENT HEADERS AND EXTENSION ATTRIBUTES BETTER
TODO: SOURCE IS ALWAYS http://localhost:8000/ WHEN COMING FROM EXECUTOR

TODO: BATCH IS BROKEN BY THIS WAY OF USING REQ IDS - NOW THE SECOND ROW OVERWRITES THE FIRST. MAY HAVE TO ADD ORDINAL TO DOC ID AND ENSURE ORIGINAL ID IN DOC BODY.
TODO: HANDLE GRPC AND INFERENCESERVICES
TODO: THINK ABOUT SHADOW CASE - HOW TO ENSURE WE HAVE SOMETHING TO LINK DEFAULT AND SHADOW? MULTIPLE SHADOWS? https://github.com/SeldonIO/seldon-core/issues/1207
TODO: FEEDBACK - IF SENT WITH CUSTOM ID HEADER COULD SUPPORT A/B TESTS WITH RECORDED RESULTS
