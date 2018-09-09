const argparse = require("argparse");
const express = require("express");
const app = express();
const bodyParser = require("body-parser");
app.use(bodyParser.urlencoded({ extended: true }));
let predict = null;
loadModel = async function(model) {
  model = "./model/" + model;

  try {
    const MyModel = require(model);
    console.log("Loading Model", model);
    let x = new MyModel();
    await x.init();
    return x.predict.bind(x);
  } catch (msg) {
    return null;
  }
};

const parser = new argparse.ArgumentParser({
  description: "Seldon-core nodejs microservice builder",
  addHelp: true
});
parser.addArgument("--model", {
  type: "string",
  help: "Name of the model File"
});
parser.addArgument("--api", {
  type: "string",
  help: "Endpoints to be exposed REST or GRPC"
});
parser.addArgument("--service", {
  type: "string",
  help: "Service type"
});
parser.addArgument("--persistence", {
  type: "int",
  defaultValue: 0,
  help: "Persistence present or not"
});
const args = parser.parseArgs();

console.log(args.model, args.api, args.service, args.persistence);

if (args.service === "MODEL" && args.api === "REST") {
  app.post("/predict", async (req, res) => {
    try {
      body = JSON.parse(req.body.json);
      body = body.data;
    } catch (msg) {
      res.status(500).send("Cannot parse predict input json " + req.body.json);
    }
    if (predict && typeof predict === "function") {
      result = predict(body.tensor, body.names);
      res.status(200).send(result);
    } else {
      res.status(500).send(args.model + " Not Found");
    }
  });
}

app.listen(5000, async () => {
  predict = await loadModel(
    args.model,
    args.api,
    args.service,
    args.persistence
  );
  console.log("NodeJs Microservice listening on port 5000!");
});
