library(methods)

predict.f <- function(f,newdata=list()) {
  predict(f$model, newdata = newdata)
}

new_f <- function(filename) {
  model <- readRDS(filename)
  structure(list(model=model), class = "f")
}

initialise_seldon <- function(params) {
  new_f("model.Rds")
}