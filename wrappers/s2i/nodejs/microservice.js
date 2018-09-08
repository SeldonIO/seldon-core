const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");
const MyModel = require("./MyModel");

let x = new MyModel();
x.init().then(() => {
  x.predict(tf.randomNormal([1, 10])).print();
});
