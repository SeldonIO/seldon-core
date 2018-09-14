const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");
const path = require("path");
const model_path = "/model.json";

class MnistClassifier {
  async init() {
    this.model = await tf.loadModel(
      "file://" + path.join(__dirname, model_path)
    );
    const optimizer = "rmsprop";
    this.model.compile({
      optimizer: optimizer,
      loss: "categoricalCrossentropy",
      metrics: ["accuracy"]
    });
  }

  predict(X, feature_names) {
    console.log("Predicting ...");
    try {
      X = tf.tensor(X);
    } catch (msg) {
      console.log("Predict input may be a Tensor already");
    }
    const result = this.model.predict(X);
    let obj = result.dataSync();
    let values = Object.keys(obj).map(key => obj[key]);
    var newValues = [];
    while (values.length) newValues.push(values.splice(0, 10));
    return newValues;
  }
}

module.exports = MnistClassifier;
