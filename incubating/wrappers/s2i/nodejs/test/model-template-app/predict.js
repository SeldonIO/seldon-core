async function loadModel(model) {
  try {
    const MyModel = require(model);
    let x = new MyModel();
    await x.init();
    return x.predict.bind(x);
  } catch (msg) {
    return null;
  }
}

async function run() {
  let predict = await loadModel("./MyModel");
  predict = predict([
    [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    [11, 2, 3, 4, 5, 6, 7, 8, 9, 11]
  ]);
  console.log(JSON.stringify(predict));
}

run();
