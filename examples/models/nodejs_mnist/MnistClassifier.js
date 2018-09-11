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
    let image = tf.tensor4d(X[0], [1, 28, 28, 1]);
    const result = this.model.predict(image);
    let obj = result.dataSync();
    let values = Object.keys(obj).map(key => obj[key]);
    return values;
  }
}

module.exports = MnistClassifier;
