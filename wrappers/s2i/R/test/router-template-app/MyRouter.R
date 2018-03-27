library(methods)


send_feedback.f <- function(f,request=list(),reward=1,truth=NULL) {
  write("MyRouter Feedback called",stdout())
}

route.f <- function(f,data) {
  write("MyRouter route called",stdout())
  0
}

new_f <- function() {
  structure(list(), class = "f")
}


initialise_seldon <- function(params) {
  new_f()
}

