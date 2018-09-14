const tf = require("@tensorflow/tfjs");
require("@tensorflow/tfjs-node");
const argparse = require("argparse");
const ProgressBar = require("progress");
const path = require("path");
const data = require("./data");
const model = require("./model_arch");

async function run(epochs, batchSize, modelSavePath) {
  await data.loadData();

  const { images: trainImages, labels: trainLabels } = data.getTrainData();
  model.summary();

  let progressBar;
  let epochBeginTime;
  let millisPerStep;
  const validationSplit = 0.15;
  const numTrainExamplesPerEpoch = trainImages.shape[0] * (1 - validationSplit);
  const numTrainBatchesPerEpoch = Math.ceil(
    numTrainExamplesPerEpoch / batchSize
  );
  await model.fit(trainImages, trainLabels, {
    epochs,
    batchSize,
    validationSplit,
    callbacks: {
      onEpochBegin: async epoch => {
        progressBar = new ProgressBar(":bar: :eta", {
          total: numTrainBatchesPerEpoch,
          head: `>`
        });
        console.log(`Epoch ${epoch + 1} / ${epochs}`);
        epochBeginTime = tf.util.now();
      },
      onBatchEnd: async (batch, logs) => {
        if (batch === numTrainBatchesPerEpoch - 1) {
          millisPerStep =
            (tf.util.now() - epochBeginTime) / numTrainExamplesPerEpoch;
        }
        progressBar.tick();
        await tf.nextFrame();
      },
      onEpochEnd: async (epoch, logs) => {
        console.log(
          `Loss: ${logs.loss.toFixed(3)} (train), ` +
            `${logs.val_loss.toFixed(3)} (val); ` +
            `Accuracy: ${logs.acc.toFixed(3)} (train), ` +
            `${logs.val_acc.toFixed(3)} (val) ` +
            `(${millisPerStep.toFixed(2)} ms/step)`
        );
        await tf.nextFrame();
      }
    }
  });

  const { images: testImages, labels: testLabels } = data.getTestData();
  const evalOutput = model.evaluate(testImages, testLabels);

  console.log(
    `\nEvaluation result:\n` +
      `  Loss = ${evalOutput[0].dataSync()[0].toFixed(3)}; ` +
      `Accuracy = ${evalOutput[1].dataSync()[0].toFixed(3)}`
  );

  if (modelSavePath != null) {
    await model.save(`file://${modelSavePath}`);
    console.log(`Saved model to path: ${modelSavePath}`);
  }
}

const parser = new argparse.ArgumentParser({
  description: "TensorFlow.js-Node MNIST Example.",
  addHelp: true
});
parser.addArgument("--epochs", {
  type: "int",
  defaultValue: 20,
  help: "Number of epochs to train the model for."
});
parser.addArgument("--batch_size", {
  type: "int",
  defaultValue: 128,
  help: "Batch size to be used during model training."
});
parser.addArgument("--model_save_path", {
  type: "string",
  help: "Path to which the model will be saved after training."
});
const args = parser.parseArgs();

console.log(
  args.epochs,
  args.batch_size,
  path.join(__dirname, args.model_save_path)
);

run(args.epochs, args.batch_size, path.join(__dirname, args.model_save_path));
