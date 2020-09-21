from flask import Flask, request, Response
import logging
import os
import threading
from elasticsearch import Elasticsearch

from seldon_core.metrics import SeldonMetrics

METRICS_ENDPOINT = os.environ.get("PREDICTIVE_UNIT_METRICS_ENDPOINT", "/metrics")
INDEX_NAME = "inference-log-seldon-seldon-sklearn-default"
elastic_host = os.getenv("ELASTICSEARCH_HOST", "localhost")
elastic_port = os.getenv("ELASTICSEARCH_PORT", 9200)

logger = logging.getLogger(__name__)

app = Flask(__name__, static_url_path="")

es = Elasticsearch(f"http://{elastic_host}:{elastic_port}/")

seldon_metrics = SeldonMetrics(worker_id_func=lambda: threading.current_thread().name)


def _calculate_initial_metrics() -> (int, int):
    query = es.search(
        index=INDEX_NAME, body={"query": {"match_all": {}}}, params={"size": 1000}
    )
    total = 0
    correct = 0
    for elem in query.get("hits", {}).get("hits", []):
        feedback = elem.get("_source", {}).get("feedback", {})
        if not feedback:
            continue
        total += 1
        if feedback.get("reward", 0) > 0:
            correct += 1

    return (total, correct)


def _add_counter(total: int, correct: int) -> None:
    metrics = []
    if total:
        metrics.append({"type": "COUNTER", "key": "total", "value": total})
    if correct:
        metrics.append({"type": "COUNTER", "key": "correct", "value": correct})
    seldon_metrics.update(metrics)


@app.route("/", methods=["GET", "POST"])
def index():
    body = request.get_json(force=True)
    # Currently we don't count requests unless they have reward
    # TODO: reconsider as we may want to add extra field(s) to reward
    if not isinstance(body, dict) or "reward" not in body:
        print("Error")
        return Response("", 500)
    reward = int(body["reward"]) > 0
    # We now add one to total, and add one if correct
    _add_counter(1, reward)
    return Response("", status=200)


@app.route(METRICS_ENDPOINT, methods=["GET"])
def metrics():
    logger.debug("REST Metrics Request")
    metrics, mimetype = seldon_metrics.generate_metrics()
    return Response(metrics, mimetype=mimetype)

if __name__ == "__main__":
    (total, correct) = _calculate_initial_metrics()
    _add_counter(total, correct)
    app.run(host="0.0.0.0", port=8080)

