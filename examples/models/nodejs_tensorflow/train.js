const tf = require("@tensorflow/tfjs");
const model_path = "";
const path = require("path");
// Load the binding:
require("@tensorflow/tfjs-node"); // Use '@tensorflow/tfjs-node-gpu' if running with GPU.

// Train a simple model:
const model = tf.sequential();
model.add(
  tf.layers.dense({ units: 100, activation: "relu", inputShape: [10] })
);
model.add(tf.layers.dense({ units: 1, activation: "linear" }));
model.compile({ optimizer: "sgd", loss: "meanSquaredError" });

const xs = tf.randomNormal([100, 10]);
const ys = tf.randomNormal([100, 1]);

async function train() {
  await model.fit(xs, ys, {
    epochs: 100,
    callbacks: {
      onEpochEnd: async (epoch, log) => {
        console.log(`Epoch ${epoch}: loss = ${log.loss}`);
      }
    }
  });
  console.log(path.join(__dirname, model_path));
  await model.save(`file://${path.join(__dirname, model_path)}`);
}

train();
