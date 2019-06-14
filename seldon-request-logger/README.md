# Example Request Logger

For use with request logging example (/examples/centralised-logging/request-logging/).

Intended to be called via knative eventing. Can also be called as regular flask http service.

Request logger component should take request, transform as appropriate and log to stdout for collection by fluentd.

Custom request loggers can be built for different types of transformations.