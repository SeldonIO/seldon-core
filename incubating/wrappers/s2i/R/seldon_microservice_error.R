seldon_microservice_error <- function(message, status_code, reason="MICROSERVICE_BAD_DATA") {
  err <- structure(
    list(message = message, status_code = status_code, reason = reason),
    class = c("seldon_microservice_error", "error", "condition")
  )
  signalCondition(err)
}
