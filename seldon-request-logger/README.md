# Example Request Logger

For use with request logging example (/examples/centralised-logging/request-logging/). The deployment yaml for this is there.

Intended to be called via knative eventing. Can also be called as regular flask http service.

Request logger component should take request, transform as appropriate and log to stdout for collection by fluentd.

Custom request loggers can be built for different types of transformations. Can be written in any language, just needs to handle HTTP POST requests and log to stdout for fluentd or could go direct to chosen backend.

Initial version just dealing with ndarray - not image or tensor.