/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package schema

import (
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	EnvURL            = "SCHEMA_REGISTRY_URL"
	EnvUsername       = "SCHEMA_REGISTRY_USERNAME"
	EnvPassword       = "SCHEMA_REGISTRY_PASSWORD"
	EnvToken          = "SCHEMA_REGISTRY_TOKEN"
	EnvTargetSr       = "SCHEMA_REGISTRY_TARGET_SR"
	EnvIdentityPoolID = "SCHEMA_REGISTRY_IDENTITY_POOL_ID"
)

func NewSchemaRegistryClient(log *log.Logger) schemaregistry.Client {
	url := os.Getenv(EnvURL)
	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)
	bearerToken := os.Getenv(EnvToken)
	targetSr := os.Getenv(EnvTargetSr)
	identityPoolID := os.Getenv(EnvIdentityPoolID)
	logger := log.WithField("func", "setup")

	if url == "" {
		return nil
	}

	var conf *schemaregistry.Config
	if bearerToken != "" {
		conf = schemaregistry.NewConfigWithBearerAuthentication(url, bearerToken, targetSr, identityPoolID)
	} else {
		logger.Info("registering with basic auth")
		conf = schemaregistry.NewConfigWithBasicAuthentication(url, username, password)
		logger.Info("registering with basic auth", "username", username, "password", password)
	}

	srClient, err := schemaregistry.NewClient(conf)
	if err != nil {
		logger.Warnf("unable to create schema registry client: %v", err)
		return nil
	}

	_, err = srClient.GetAllSubjects()
	if err != nil {
		logger.Warnf("unable to get all subjects: %v", err)
	}

	srClient.Config()

	logger.Info("schema registry client created")
	return srClient
}

func GetModelInferenceRequestSchema() schemaregistry.SchemaInfo {
	protoSchema := `
syntax = "proto3";
package inference;

message ModelInferRequest
{
  // The data contained in a tensor represented by the repeated type
  // that matches the tensor's data type. Protobuf oneof is not used
  // because oneofs cannot contain repeated fields.
  message InferTensorContents
  {
    // Representation for BOOL data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated bool bool_contents = 1;

    // Representation for INT8, INT16, and INT32 data types. The size
    // must match what is expected by the tensor's shape. The contents
    // must be the flattened, one-dimensional, row-major order of the
    // tensor elements.
    repeated int32 int_contents = 2;

    // Representation for INT64 data types. The size must match what
    // is expected by the tensor's shape. The contents must be the
    // flattened, one-dimensional, row-major order of the tensor elements.
    repeated int64 int64_contents = 3;

    // Representation for UINT8, UINT16, and UINT32 data types. The size
    // must match what is expected by the tensor's shape. The contents
    // must be the flattened, one-dimensional, row-major order of the
    // tensor elements.
    repeated uint32 uint_contents = 4;

    // Representation for UINT64 data types. The size must match what
    // is expected by the tensor's shape. The contents must be the
    // flattened, one-dimensional, row-major order of the tensor elements.
    repeated uint64 uint64_contents = 5;

    // Representation for FP32 data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated float fp32_contents = 6;

    // Representation for FP64 data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated double fp64_contents = 7;

    // Representation for BYTES data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated bytes bytes_contents = 8;
  }

  // An inference parameter value. The Parameters message describes a
  // “name”/”value” pair, where the “name” is the name of the parameter
  // and the “value” is a boolean, integer, or string corresponding to
  // the parameter.
  message InferParameter
  {
    // The parameter value can be a string, an int64, a boolean
    // or a message specific to a predefined parameter.
    oneof parameter_choice
    {
      // A boolean parameter value.
      bool bool_param = 1;

      // An int64 parameter value.
      int64 int64_param = 2;

      // A string parameter value.
      string string_param = 3;
    }
  }

  // An input tensor for an inference request.
  message InferInputTensor
  {
    // The tensor name.
    string name = 1;

    // The tensor data type.
    string datatype = 2;

    // The tensor shape.
    repeated int64 shape = 3;

    // Optional inference input tensor parameters.
    map<string, InferParameter> parameters = 4;

    // The tensor contents using a data-type format. This field must
    // not be specified if "raw" tensor contents are being used for
    // the inference request.
    InferTensorContents contents = 5;
  }

  // An output tensor requested for an inference request.
  message InferRequestedOutputTensor
  {
    // The tensor name.
    string name = 1;

    // Optional requested output tensor parameters.
    map<string, InferParameter> parameters = 2;
  }

  // The name of the model to use for inferencing.
  string model_name = 1;

  // The version of the model to use for inference. If not given the
  // server will choose a version based on the model and internal policy.
  string model_version = 2;

  // Optional identifier for the request. If specified will be
  // returned in the response.
  string id = 3;

  // Optional inference parameters.
  map<string, InferParameter> parameters = 4;

  // The input tensors for the inference.
  repeated InferInputTensor inputs = 5;

  // The requested output tensors for the inference. Optional, if not
  // specified all outputs produced by the model will be returned.
  repeated InferRequestedOutputTensor outputs = 6;

  // The data contained in an input tensor can be represented in "raw"
  // bytes form or in the repeated type that matches the tensor's data
  // type. To use the raw representation 'raw_input_contents' must be
  // initialized with data for each tensor in the same order as
  // 'inputs'. For each tensor, the size of this content must match
  // what is expected by the tensor's shape and data type. The raw
  // data must be the flattened, one-dimensional, row-major order of
  // the tensor elements without any stride or padding between the
  // elements. Note that the FP16 data type must be represented as raw
  // content as there is no specific data type for a 16-bit float
  // type.
  //
  // If this field is specified then InferInputTensor::contents must
  // not be specified for any input tensor.
  repeated bytes raw_input_contents = 7;
}`

	return schemaregistry.SchemaInfo{
		Schema:     protoSchema,
		SchemaType: "PROTOBUF",
	}
}

func GetModelInferenceResponseSchema() schemaregistry.SchemaInfo {
	protoSchema := `
syntax = "proto3";
package inference;

message ModelInferResponse
{
  // The data contained in a tensor represented by the repeated type
  // that matches the tensor's data type. Protobuf oneof is not used
  // because oneofs cannot contain repeated fields.
  message InferTensorContents
  {
    // Representation for BOOL data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated bool bool_contents = 1;

    // Representation for INT8, INT16, and INT32 data types. The size
    // must match what is expected by the tensor's shape. The contents
    // must be the flattened, one-dimensional, row-major order of the
    // tensor elements.
    repeated int32 int_contents = 2;

    // Representation for INT64 data types. The size must match what
    // is expected by the tensor's shape. The contents must be the
    // flattened, one-dimensional, row-major order of the tensor elements.
    repeated int64 int64_contents = 3;

    // Representation for UINT8, UINT16, and UINT32 data types. The size
    // must match what is expected by the tensor's shape. The contents
    // must be the flattened, one-dimensional, row-major order of the
    // tensor elements.
    repeated uint32 uint_contents = 4;

    // Representation for UINT64 data types. The size must match what
    // is expected by the tensor's shape. The contents must be the
    // flattened, one-dimensional, row-major order of the tensor elements.
    repeated uint64 uint64_contents = 5;

    // Representation for FP32 data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated float fp32_contents = 6;

    // Representation for FP64 data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated double fp64_contents = 7;

    // Representation for BYTES data type. The size must match what is
    // expected by the tensor's shape. The contents must be the flattened,
    // one-dimensional, row-major order of the tensor elements.
    repeated bytes bytes_contents = 8;
  }

  // An inference parameter value. The Parameters message describes a
  // “name”/”value” pair, where the “name” is the name of the parameter
  // and the “value” is a boolean, integer, or string corresponding to
  // the parameter.
  message InferParameter
  {
    // The parameter value can be a string, an int64, a boolean
    // or a message specific to a predefined parameter.
    oneof parameter_choice
    {
      // A boolean parameter value.
      bool bool_param = 1;

      // An int64 parameter value.
      int64 int64_param = 2;

      // A string parameter value.
      string string_param = 3;
    }
  }

  // An output tensor returned for an inference request.
  message InferOutputTensor
  {
    // The tensor name.
    string name = 1;

    // The tensor data type.
    string datatype = 2;

    // The tensor shape.
    repeated int64 shape = 3;

    // Optional output tensor parameters.
    map<string, InferParameter> parameters = 4;

    // The tensor contents using a data-type format. This field must
    // not be specified if "raw" tensor contents are being used for
    // the inference response.
    InferTensorContents contents = 5;
  }

  // The name of the model used for inference.
  string model_name = 1;

  // The version of the model used for inference.
  string model_version = 2;

  // The id of the inference request if one was specified.
  string id = 3;


  // Optional inference response parameters.
  map<string, InferParameter> parameters = 4;

  // The output tensors holding inference results.
  repeated InferOutputTensor outputs = 5;

  // The data contained in an output tensor can be represented in
  // "raw" bytes form or in the repeated type that matches the
  // tensor's data type. To use the raw representation 'raw_output_contents'
  // must be initialized with data for each tensor in the same order as
  // 'outputs'. For each tensor, the size of this content must match
  // what is expected by the tensor's shape and data type. The raw
  // data must be the flattened, one-dimensional, row-major order of
  // the tensor elements without any stride or padding between the
  // elements. Note that the FP16 data type must be represented as raw
  // content as there is no specific data type for a 16-bit float
  // type.
  //
  // If this field is specified then InferOutputTensor::contents must
  // not be specified for any output tensor.
  repeated bytes raw_output_contents = 6;
}`

	return schemaregistry.SchemaInfo{
		Schema:     protoSchema,
		SchemaType: "PROTOBUF",
	}
}
