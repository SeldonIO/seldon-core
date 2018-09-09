const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");
const path = require("path");
const model_path = "/model.json";

let MyModel = function() {};

MyModel.prototype.init = async function() {
  this.model = await tf.loadModel("file://" + path.join(__dirname, model_path));
  this.model.compile({ optimizer: "sgd", loss: "meanSquaredError" });
};

MyModel.prototype.predict = function(data, names) {
  console.log("Predicting ...");
  let result = this.model.predict(tf.tensor(data.values, data.shape));
  let obj = result.dataSync();
  let values = Object.keys(obj).map(key => obj[key]);
  return {
    data: { names: [], tensor: { shape: result.shape, values: values } }
  };
};

module.exports = MyModel;
