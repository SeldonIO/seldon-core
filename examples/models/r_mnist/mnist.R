library(methods)

predict.mnist <- function(mnist,newdata=list()) {
  predict(mnist$model, newdata = newdata, type='prob')
}

new_mnist <- function(filename) {
  model <- readRDS(filename)
  structure(list(model=model), class = "mnist")
}

initialise_seldon <- function(params) {
  new_mnist("model.Rds")
}