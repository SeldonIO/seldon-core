library(plumber)
library(jsonlite)
library("optparse")
library(methods)

v <- function(...) cat(sprintf(...), sep='', file=stdout())

validate_json <- function(jdf) {
  if (!"data" %in% names(jdf)) {
    return("data field is missing")
  }
  else if (!("ndarray" %in% names(jdf$data) || "tensor" %in% names(jdf$data)) ) {
    return("data field must contain ndarray or tensor field")
  }
  else{
    return("OK")
  }
}

extract_data <- function(jdf) {
  if ("ndarray" %in% names(jdf$data)){
    jdf$data$ndarray
  } else {
    data <- jdf$data$tensor$values
    dim(data) <- jdf$data$tensor$shape
    data
  }
}

extract_names <- function(jdf) {
  if ("names" %in% names(jdf$data)) {
    jdf$data$names
  } else {
    list()
  }
}

create_response <- function(req_df,res_df){
  if ("ndarray" %in% names(req_df$data)){
    templ <- '{"data":{"names":%s,"ndarray":%s}}'
    names <- toJSON(colnames(res_df))
    values <- toJSON(as.matrix(res_df))
    sprintf(templ,names,values)
  } else {
    templ <- '{"data":{"names":%s,"tensor":{"shape":%s,"values":%s}}}'
    names <- toJSON(colnames(res_df))
    values <- toJSON(c(res_df))
    dims <- toJSON(dim(res_df))
    sprintf(templ,names,dims,values)
  }
}


predict_endpoint <- function(req,res,json=NULL,isDefault=NULL) {
  write("Predict called", stdout())
  jdf <- fromJSON(json)
  valid_input <- validate_json(jdf)
  if (valid_input[1] == "OK") {
    data = extract_data(jdf)
    names = extract_names(jdf)
    df <- data.frame(data)
    colnames(df) <- names
    scores <- predict(user_model,newdata=df)
    res_json = create_response(jdf,scores)
    res$body <- res_json
    res
  } else {
    res$status <- 400 # Bad request
    list(error=jsonlite::unbox(valid_input))
  }
}

parse_commandline <- function() {
  parser <- OptionParser()
  parser <- add_option(parser, c("-p", "--parameters"), type="character",
                       help="Parameters for component", metavar = "parameters")
  parser <- add_option(parser, c("-m", "--model"), type="character",
                       help="Model file", metavar = "model")
  parser <- add_option(parser, c("-s", "--service"), type="character",
                       help="Service type", metavar = "service", default = "MODEL")
  parser <- add_option(parser, c("-a", "--api"), type="character",
                       help="API type - REST", metavar = "api", default = "REST")
  args <- parse_args(parser, args = commandArgs(trailingOnly = TRUE),
                     convert_hyphens_to_underscores = TRUE)
  
  if (is.null(args$parameters)){
    args$parameters <- Sys.getenv("PREDICTIVE_UNIT_PARAMETERS")
  }
  
  if (args$parameters == ''){
    args$parameters = "[]"
  }
  
  args
}


extract_parmeters <- function(params) {
  j = fromJSON(params)
  values <- list()
  names <- list()
  for (i in seq_along(j))
  {
    name <- j[i,"name"]
    value <- j[i,"value"]
    type <- j[i,"type"]
    if (type == "INT")
      value <- as.integer(value)
    else if (type == "FLOAT")
      value <- as.double(value)
    else if (type == "BOOL")
      value <- as.logical(type.convert(value))
    values <- c(values,value)
    names <- c(names,name)
  }
  names(values) <- names
  values
}

validate_commandline <- function(args) {
  if (!is.element(args$service,c("MODEL","ROUTER","COMBINER","TRANSFORMER"))) {
    v("Invalid service type [%s]\n",args$service)
    1
  }else if (!is.element(args$api,c("REST"))) {
    v("Invalid API type [%s]\n",args$api)
    1
  }
  else{
    0
  }
}

# Parse command line and validate
args <- parse_commandline()
if (validate_commandline(args) > 0){
  quit(status=1)
}
params <- extract_parmeters(args$parameters)

# Check user model exists
if(!file.exists(args$model)){
  v("Model file does not exist [%s]\n",args$model)
  quit(status=1)
}

#Load user model
source(args$model)
user_model <- initialise_seldon(params)

# Setup generics
send_feedback <- function(x,...) UseMethod("send_feedback", x)
route <- function(x,...) UseMethod("route",x)


serve_model <- plumber$new()
if (args$service == "MODEL") {
  serve_model$handle("POST", "/predict",predict_endpoint)
  serve_model$handle("GET", "/predict",predict_endpoint)
} else
{
  v("Unknown service type [%s]\n",args$service)
  quit(status=1)
}
serve_model$run(port = 8000)
