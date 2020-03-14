library(methods)

predict.iris <- function(iris,newdata=list()) {
  predict(iris$model, newdata = newdata)
}

new_iris <- function(filename) {
  model <- readRDS(filename)
  structure(list(model=model), class = "iris")
}

initialise_seldon <- function(params) {
  new_iris("model.Rds")
}