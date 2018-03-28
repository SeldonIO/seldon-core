library(methods)

predict.mymodel <- function(mymodel,newdata=list()) {
  write("MyModel predict called", stdout())
  newdata
}


new_mymodel <- function() {
  structure(list(), class = "mymodel")
}


initialise_seldon <- function(params) {
  new_mymodel()
}