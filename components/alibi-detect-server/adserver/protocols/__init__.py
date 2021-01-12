from enum import Enum


class Protocol(Enum):
    tensorflow_http = "tensorflow.http"
    seldon_http = "seldon.http"
    seldonfeedback_http = "seldonfeedback.http"
    v2_http = "v2.http"

    def __str__(self):
        return self.value
