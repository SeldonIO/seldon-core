const express = require("express");
const app = express();
const bodyParser = require("body-parser");
const grpc = require("grpc");
const grpc_services = require("./prediction_grpc_pb");

module.exports = (
  user_model,
  api,
  port,
  rest_data_to_array,
  array_to_rest_data,
  array_to_grpc_data
) => {
  if (user_model.predict && typeof user_model.predict === "function") {
    console.log("Predict function loaded successfully");
  } else if (
    user_model.send_feedback &&
    typeof user_model.send_feedback === "function"
  ) {
    console.log("Send feedback function loaded successfully");
  } else {
    console.log(user_model);
    process.exit(1);
  }
  let predict = user_model.predict ? user_model.predict.bind(user_model) : null;
  let send_feedback = user_model.send_feedback
    ? user_model.send_feedback.bind(user_model)
    : null;

  if (api === "REST") {
    app.use(bodyParser.urlencoded({ extended: true }));
    app.post("/predict", (req, res) => {
      try {
        body = JSON.parse(req.body.json);
        body = body.data;
      } catch (msg) {
        console.log(msg);
        res.status(500).send("Cannot parse predict input json " + req.body);
      }
      if (predict && typeof predict === "function") {
        result = predict(rest_data_to_array(body), body.names);
        result = { data: array_to_rest_data(result, body) };
        res.status(200).send(result);
      } else {
        console.log("Predict function not Found");
        res.status(500).send(null);
      }
    });
    app.post("/send-feedback", (req, res) => {
      try {
        body = JSON.parse(req.body.json);
        request = body.request;
        body = request.data;
      } catch (msg) {
        console.log(msg);
        res
          .status(500)
          .send("Cannot parse Send feedback input json " + req.body);
      }
      if (send_feedback && typeof send_feedback === "function") {
        result = send_feedback(
          rest_data_to_array(body),
          body.names,
          rest_data_to_array(request.truth),
          request.reward
        );
        // result = { data: array_to_rest_data(result, body) };
        res.status(200).send({});
      } else {
        console.log("Send feedback function not Found");
        res.status(500).send(null);
      }
    });
    var server = app.listen(port, () => {
      console.log(`NodeJs REST Microservice listening on port ${port}!`);
    });
    function stopServer(code) {
      server.close();
      console.log(`About to exit with code: ${code}`);
    }
    process.on("SIGINT", stopServer.bind(this));
    process.on("SIGTERM", stopServer.bind(this));
  }

  if (api === "GRPC") {
    function predictEndpoint(call, callback) {
      let data = call.request.getData();
      let body = { names: data.getNamesList() };

      if (data.hasTensor()) {
        data = data.getTensor();
        body["tensor"] = {
          shape: data.getShapeList(),
          values: data.getValuesList()
        };
      } else {
        body["ndarray"] = data.getNdarray();
      }
      result = predict(rest_data_to_array(body), body.names);
      callback(null, array_to_grpc_data(result, body));
    }
    function feedbackEndpoint(call, callback) {
      let request = call.request.getRequest();
      let data = request.getData();
      let body = { names: data.getNamesList() };

      if (data.hasTensor()) {
        data = data.getTensor();
        body["tensor"] = {
          shape: data.getShapeList(),
          values: data.getValuesList()
        };
      } else {
        body["ndarray"] = data.getNdarray();
      }

      let truth = call.request.getTruth();
      let truth_data = truth.getData();
      let truth_body = { names: truth_data.getNamesList() };

      if (truth_data.hasTensor()) {
        truth_data = truth_data.getTensor();
        truth_body["tensor"] = {
          shape: truth_data.getShapeList(),
          values: truth_data.getValuesList()
        };
      } else {
        truth_body["ndarray"] = truth_data.getNdarray();
      }
      result = send_feedback(
        rest_data_to_array(body),
        body.names,
        rest_data_to_array(truth_body),
        call.request.getReward()
      );
      callback(null, {});
    }
    var server = new grpc.Server();
    server.addService(grpc_services.ModelService, { predict: predictEndpoint });
    // server.addService(grpc_services.SeldonService, {
    //   predict: predictEndpoint,
    //   sendFeedback: feedbackEndpoint
    // });
    server.bind("0.0.0.0:" + port, grpc.ServerCredentials.createInsecure());
    server.start();
    console.log(`NodeJs GRPC Microservice listening on port ${port}!`);
    function stopServer(code) {
      server.forceShutdown();
      console.log(`About to exit with code: ${code}`);
    }
    process.on("SIGINT", stopServer.bind(this));
    process.on("SIGTERM", stopServer.bind(this));
  }
};
