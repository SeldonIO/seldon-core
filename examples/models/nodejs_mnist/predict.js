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
  result = predict([
    testImages
      .flatten()
      .slice([0], [784])
      .dataSync()
  ]);
  console.log("Predicted Result\n", result);
  console.log(
    "Actual Result\n",
    testLabels
      .flatten()
      .slice([0], [10])
      .dataSync()
  );
}

run();
