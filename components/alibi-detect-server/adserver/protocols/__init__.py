from enum import Enum


class Protocol(Enum):
    tensorflow_http = "tensorflow.http"
    seldon_http = "seldon.http"
    seldonfeedback_http = "seldonfeedback.http"
    kfserving_http = "kfserving.http"

    def __str__(self):
        return self.value
