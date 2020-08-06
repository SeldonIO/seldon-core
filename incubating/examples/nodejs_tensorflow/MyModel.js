const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");
const path = require("path");
const model_path = "/model.json";

let MyModel = function() {};

MyModel.prototype.init = async function() {
  this.model = await tf.loadModel("file://" + path.join(__dirname, model_path));
  this.model.compile({ optimizer: "sgd", loss: "meanSquaredError" });
};

MyModel.prototype.predict = function(X, names) {
  console.log("Predicting ...");
  try {
    X = tf.tensor(X);
  } catch (msg) {
    console.log("Predict input may be a Tensor already");
  }
  let result = this.model.predict(X);
  let obj = result.dataSync();
  let values = Object.keys(obj).map(key => obj[key]);
  var newValues = [];
  while (values.length) newValues.push(values.splice(0, 1));
  return newValues;
};

module.exports = MyModel;
