const data = require("./data");
const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");

async function loadModel(model) {
  try {
    const MnistClassifier = require(model);
    let x = new MnistClassifier();
    await x.init();
    return x.predict.bind(x);
  } catch (msg) {
    return null;
  }
}

async function run() {
  await data.loadData();
  const { images: testImages, labels: testLabels } = data.getTestData();
  let predict = await loadModel("./MnistClassifier");
  result = predict(testImages);
  console.log("Predicted Result Size\n", result.length);
  console.log("Actual Result Size\n", testLabels.dataSync().length);
}

run();
