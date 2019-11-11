let MyTransformer = function() {};

MyTransformer.prototype.init = function() {
  console.log("Initializing Transform ...");
};

MyTransformer.prototype.transform_input = function(X, names) {
  console.log("Identity Transform ...");
  return X;
};

MyTransformer.prototype.transform_output = function(X, names) {
  console.log("Identity Transform ...");
  return X;
};

module.exports = MyTransformer;
