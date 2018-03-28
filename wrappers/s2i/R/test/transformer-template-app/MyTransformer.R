library(methods)

transform_input.f <- function(f,newdata=list()) {
 write("Transform input", stdout())
 newdata
}

transform_output.f <- function(f,newdata=list()) {
  write("Transform output", stdout())
  newdata
}

new_f <- function() {
  structure(list(), class = "f")
}


initialise_seldon <- function(params) {
  new_f()
}

