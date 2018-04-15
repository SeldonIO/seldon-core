library(methods)

predict.mnist <- function(mnist,newdata=list()) {
  cn <- 1:784
  for (i in seq_along(cn)){cn[i] <- paste("X",cn[i],sep = "")}
  colnames(newdata) <- cn
  predict(mnist$model, newdata = newdata, type='prob')
}

new_mnist <- function(filename) {
  model <- readRDS(filename)
  structure(list(model=model), class = "mnist")
}

initialise_seldon <- function(params) {
  new_mnist("model.Rds")
}