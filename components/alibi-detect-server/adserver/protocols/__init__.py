from enum import Enum


class Protocol(Enum):
    tensorflow_http = "tensorflow.http"
    seldon_http = "seldon.http"
    seldonfeedback_http = "seldonfeedback.http"

    def __str__(self):
        return self.value
