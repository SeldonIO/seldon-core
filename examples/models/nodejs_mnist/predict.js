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
  let predict = await loadModel("./MnistClassifier");
  result = predict([
    Array(784)
      .fill(0)
      .map(() => Math.random())
  ]);
  console.log(result);
}

run();
