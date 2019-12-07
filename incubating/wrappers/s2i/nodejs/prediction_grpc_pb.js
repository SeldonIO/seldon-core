// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var prediction_pb = require('./prediction_pb.js');
var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js');

function serialize_seldon_protos_Feedback(arg) {
  if (!(arg instanceof prediction_pb.Feedback)) {
    throw new Error('Expected argument of type seldon.protos.Feedback');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_seldon_protos_Feedback(buffer_arg) {
  return prediction_pb.Feedback.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_seldon_protos_SeldonMessage(arg) {
  if (!(arg instanceof prediction_pb.SeldonMessage)) {
    throw new Error('Expected argument of type seldon.protos.SeldonMessage');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_seldon_protos_SeldonMessage(buffer_arg) {
  return prediction_pb.SeldonMessage.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_seldon_protos_SeldonMessageList(arg) {
  if (!(arg instanceof prediction_pb.SeldonMessageList)) {
    throw new Error('Expected argument of type seldon.protos.SeldonMessageList');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_seldon_protos_SeldonMessageList(buffer_arg) {
  return prediction_pb.SeldonMessageList.deserializeBinary(new Uint8Array(buffer_arg));
}


// [END Messages]
//
// [START Services]
//
var GenericService = exports.GenericService = {
  transformInput: {
    path: '/seldon.protos.Generic/TransformInput',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  transformOutput: {
    path: '/seldon.protos.Generic/TransformOutput',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  route: {
    path: '/seldon.protos.Generic/Route',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  aggregate: {
    path: '/seldon.protos.Generic/Aggregate',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessageList,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessageList,
    requestDeserialize: deserialize_seldon_protos_SeldonMessageList,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  sendFeedback: {
    path: '/seldon.protos.Generic/SendFeedback',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.Feedback,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_Feedback,
    requestDeserialize: deserialize_seldon_protos_Feedback,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.GenericClient = grpc.makeGenericClientConstructor(GenericService);
var ModelService = exports.ModelService = {
  predict: {
    path: '/seldon.protos.Model/Predict',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.ModelClient = grpc.makeGenericClientConstructor(ModelService);
var RouterService = exports.RouterService = {
  route: {
    path: '/seldon.protos.Router/Route',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  sendFeedback: {
    path: '/seldon.protos.Router/SendFeedback',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.Feedback,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_Feedback,
    requestDeserialize: deserialize_seldon_protos_Feedback,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.RouterClient = grpc.makeGenericClientConstructor(RouterService);
var TransformerService = exports.TransformerService = {
  transformInput: {
    path: '/seldon.protos.Transformer/TransformInput',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.TransformerClient = grpc.makeGenericClientConstructor(TransformerService);
var OutputTransformerService = exports.OutputTransformerService = {
  transformOutput: {
    path: '/seldon.protos.OutputTransformer/TransformOutput',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.OutputTransformerClient = grpc.makeGenericClientConstructor(OutputTransformerService);
var CombinerService = exports.CombinerService = {
  aggregate: {
    path: '/seldon.protos.Combiner/Aggregate',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessageList,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessageList,
    requestDeserialize: deserialize_seldon_protos_SeldonMessageList,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.CombinerClient = grpc.makeGenericClientConstructor(CombinerService);
var SeldonService = exports.SeldonService = {
  predict: {
    path: '/seldon.protos.Seldon/Predict',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.SeldonMessage,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_SeldonMessage,
    requestDeserialize: deserialize_seldon_protos_SeldonMessage,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
  sendFeedback: {
    path: '/seldon.protos.Seldon/SendFeedback',
    requestStream: false,
    responseStream: false,
    requestType: prediction_pb.Feedback,
    responseType: prediction_pb.SeldonMessage,
    requestSerialize: serialize_seldon_protos_Feedback,
    requestDeserialize: deserialize_seldon_protos_Feedback,
    responseSerialize: serialize_seldon_protos_SeldonMessage,
    responseDeserialize: deserialize_seldon_protos_SeldonMessage,
  },
};

exports.SeldonClient = grpc.makeGenericClientConstructor(SeldonService);
