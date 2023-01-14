/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/status"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	SeldonModelHeader    = "seldon-model"
	SeldonRouteHeader    = "x-seldon-route"
	SeldonPipelineHeader = "pipeline"
	HeaderSeparator      = "="
)

const (
	tyBool   = "BOOL"
	tyUint8  = "UINT8"
	tyUint16 = "UINT16"
	tyUint32 = "UINT32"
	tyUint64 = "UINT64"
	tyInt8   = "INT8"
	tyInt16  = "INT16"
	tyInt32  = "INT32"
	tyInt64  = "INT64"
	tyFp16   = "FP16"
	tyFp32   = "FP32"
	tyFp64   = "FP64"
	tyBytes  = "BYTES"
)

type InferType uint32

const (
	InferModel InferType = iota
	InferPipeline
	InferExplainer
)

type InferenceClient struct {
	host        string
	callOptions []grpc.CallOption
	counts      map[string]int
	errors      map[string]map[int]int
	config      *SeldonCLIConfig
}

type V2Error struct {
	Error string `json:"error"`
}

type V2InferenceResponse struct {
	ModelName    string                 `json:"model_name,omitempty"`
	ModelVersion string                 `json:"model_version,omitempty"`
	Id           string                 `json:"id"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Outputs      []interface{}          `json:"outputs,omitempty"`
}

type V2Metadata struct {
	Name     string             `json:"name"`
	Versions []string           `json:"versions,omitempty"`
	Platform string             `json:"platform,omitempty"`
	Inputs   []V2MetadataTensor `json:"inputs,omitempty"`
	Outputs  []V2MetadataTensor `json:"outputs,omitempty"`
}

type V2MetadataTensor struct {
	Name     string `json:"name"`
	Datatype string `json:"datatype"`
	Shape    []int  `json:"shape"`
}

type LogOptions struct {
	ShowHeaders  bool
	ShowRequest  bool
	ShowResponse bool
}

type CallOptions struct {
	InferProtocol string // REST or gRPC
	InferType     InferType
	StickySession bool
	Iterations    int
	Seconds       int64
}

func NewInferenceClient(host string, hostIsSet bool) (*InferenceClient, error) {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	config, err := LoadSeldonCLIConfig()
	if err != nil {
		return nil, err
	}

	// Overwrite host if set in config
	if !hostIsSet && config.Dataplane != nil && config.Dataplane.InferHost != "" {
		host = config.Dataplane.InferHost
	}

	return &InferenceClient{
		host:        host,
		callOptions: opts,
		counts:      make(map[string]int),
		errors:      make(map[string]map[int]int),
		config:      config,
	}, nil
}

func (ic *InferenceClient) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if ic.config.Dataplane.KeyPath != "" && ic.config.Dataplane.CrtPath != "" {
		certificate, err := tls.LoadX509KeyPair(ic.config.Dataplane.CrtPath, ic.config.Dataplane.KeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}

	if ic.config.Dataplane.CaPath != "" {
		ca, err := os.ReadFile(ic.config.Dataplane.CaPath)
		if err != nil {
			return nil, err
		}

		capool := x509.NewCertPool()
		if !capool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("Failed to load ca crt from %s", ic.config.Dataplane.CaPath)
		}
		tlsConfig.RootCAs = capool
	}

	if ic.config.Dataplane.SkipSSLVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}

func (ic *InferenceClient) createGrpcTransportCredentials() (credentials.TransportCredentials, error) {
	if ic.config.Dataplane != nil && ic.config.Dataplane.Tls {
		tlsConfig, err := ic.createTLSConfig()
		if err != nil {
			return nil, err
		}
		return credentials.NewTLS(tlsConfig), nil
	} else {
		return insecure.NewCredentials(), nil
	}
}

func (ic *InferenceClient) newGRPCConnection(authority string, logOpts *LogOptions) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	creds, err := ic.createGrpcTransportCredentials()
	if err != nil {
		return nil, err
	}
	metadataLoggerUnary := getMetadataLoggingUnaryInterceptor(authority, logOpts)
	retryUnary := grpc_retry.UnaryClientInterceptor(retryOpts...)
	retryStream := grpc_retry.StreamClientInterceptor(retryOpts...)

	opts := []grpc.DialOption{
		grpc.WithAuthority(authority),
		grpc.WithTransportCredentials(creds),
		grpc.WithStreamInterceptor(retryStream),
		grpc.WithChainUnaryInterceptor(
			// Ordering is important here, as each interceptor effectively invokes the next.
			// If we used the retry interceptor before the metadata-logging one, we might see
			// the metadata being logged multiple times, which is undesirable.
			metadataLoggerUnary,
			retryUnary,
		),
	}

	conn, err := grpc.Dial(ic.host, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getMetadataLoggingUnaryInterceptor(authority string, logOpts *LogOptions) grpc.UnaryClientInterceptor {
	interceptor := func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if logOpts.ShowRequest {
			i, ok := req.(*v2_dataplane.ModelInferRequest)
			if ok {
				printProto(i)
			}
		}

		if logOpts.ShowHeaders {
			host := authority
			if host == "" {
				host = cc.Target()
			}

			fmt.Printf("> %s HTTP/2\n", method)
			fmt.Printf("> Host: %s\n", host)

			md, ok := metadata.FromOutgoingContext(ctx)
			if ok {
				for k, v := range md {
					fmt.Printf("> %s:%v\n", k, v)
				}
			}

			fmt.Println()
		}

		var headers, trailers metadata.MD
		respHeaders := grpc.Header(&headers)
		respTrailers := grpc.Trailer(&trailers)
		opts = append(opts, respHeaders, respTrailers)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return err
		}

		if logOpts.ShowHeaders {
			for k, v := range headers {
				fmt.Printf("< %s:%v\n", k, v)
			}

			for k, v := range trailers {
				fmt.Printf("<< %s:%v\n", k, v)
			}

			fmt.Println()
		}

		return nil
	}

	return interceptor
}

func (ic *InferenceClient) getUrl(path string) *url.URL {
	scheme := "http"
	if ic.config.Dataplane != nil && ic.config.Dataplane.Tls {
		scheme = "https"
	}
	return &url.URL{
		Scheme: scheme,
		Host:   ic.host,
		Path:   path,
	}
}

func decodeV2Error(response *http.Response, b []byte) error {
	if response.StatusCode == http.StatusBadRequest {
		v2Error := V2Error{}
		err := json.Unmarshal(b, &v2Error)
		if err != nil {
			return fmt.Errorf("V2 server error: %d %s", response.StatusCode, b)
		}
		return fmt.Errorf("%s", v2Error.Error)
	} else {
		return fmt.Errorf("V2 server error: %d %s", response.StatusCode, b)
	}

}

func (ic *InferenceClient) createHttpClient() (*http.Client, error) {
	if ic.config.Dataplane != nil && ic.config.Dataplane.Tls {
		tlsConfig, err := ic.createTLSConfig()
		if err != nil {
			return nil, err
		}
		t := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		client := http.Client{Transport: t, Timeout: 15 * time.Second}
		return &client, nil
	} else {
		return http.DefaultClient, nil
	}
}

func (ic *InferenceClient) httpCall(
	resourceName string,
	path string,
	data []byte,
	inferType InferType,
	showHeaders bool,
	headers []string,
	authority string,
	stickySessionKeys []string,
) ([]byte, error) {
	hs, err := validateHeaders(headers)
	if err != nil {
		return nil, err
	}

	v2Url := ic.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Close = true

	if authority != "" {
		req.Host = authority
	}

	addContentTypeToRequest(req)
	addStickySessionToRequest(req, stickySessionKeys)
	addSeldonModelHeaderToRequest(req, inferType, resourceName)
	addHeadersToRequest(req, hs)

	if showHeaders {
		fmt.Printf("> %s %s %s\n", req.Method, req.URL.Path, req.Proto)
		fmt.Printf("> Host: %s\n", req.Host)
		for k, v := range req.Header {
			fmt.Printf("> %s:%v\n", k, v)
		}
		fmt.Println()
	}

	client, err := ic.createHttpClient()
	if err != nil {
		return nil, err
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

	if showHeaders {
		for k, v := range response.Header {
			fmt.Printf("< %s:%v\n", k, v)
		}
		fmt.Println()
	}

	if response.StatusCode != http.StatusOK {
		ic.updateErrors(response.StatusCode, response.Header.Values(SeldonRouteHeader))
		return nil, decodeV2Error(response, b)
	}

	_, err = saveStickySessionKeyHttp(response.Header)
	ic.updateSummary(response.Header.Values(SeldonRouteHeader))
	if err != nil {
		return b, err
	}

	return b, nil
}

func addContentTypeToRequest(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
}

func addStickySessionToRequest(r *http.Request, stickySessionKeys []string) {
	for _, k := range stickySessionKeys {
		r.Header.Add(SeldonRouteHeader, k)
	}
}

func addSeldonModelHeaderToRequest(r *http.Request, inferType InferType, resourceName string) {
	var headerValue string
	switch inferType {
	case InferModel:
		headerValue = resourceName
	case InferPipeline:
		headerValue = resourceName + "." + SeldonPipelineHeader
	}

	r.Header.Set(SeldonModelHeader, headerValue)
}

func addHeadersToRequest(r *http.Request, headers map[string]string) {
	for k, v := range headers {
		r.Header.Set(k, v)
	}
}

func (ic *InferenceClient) updateSummary(modelNames []string) {
	for _, modelName := range modelNames {
		if count, ok := ic.counts[modelName]; ok {
			ic.counts[modelName] = count + 1
		} else {
			ic.counts[modelName] = 1
		}
	}
}

func (ic *InferenceClient) updateErrors(code int, modelNames []string) {
	if len(modelNames) == 0 {
		modelNames = append(modelNames, "")
	}
	for _, modelName := range modelNames {
		if codes, ok := ic.errors[modelName]; ok {
			if count, ok2 := codes[code]; ok2 {
				codes[code] = count + 1
			} else {
				codes = make(map[int]int)
				codes[code] = 1
				ic.errors[modelName] = codes
			}

		} else {
			ic.errors[modelName] = make(map[int]int)
			ic.errors[modelName][code] = 1
		}
	}
}

func (ic *InferenceClient) ModelMetadata(modelName string, authority string) error {
	path := fmt.Sprintf("/v2/models/%s", modelName)
	v2Url := ic.getUrl(path)
	req, err := http.NewRequest("GET", v2Url.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set(SeldonModelHeader, modelName)
	req.Host = authority

	client, err := ic.createHttpClient()
	if err != nil {
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return decodeV2Error(response, b)
	}
	printPrettyJson(b)
	return nil
}

func canContinueInfer(callOptions *CallOptions, i int, timeStart int64) bool {
	return (callOptions.Seconds == 0 && i < callOptions.Iterations) ||
		(callOptions.Seconds > 0 && time.Now().Unix()-timeStart < callOptions.Seconds)
}

func (ic *InferenceClient) InferRest(
	resourceName string,
	data []byte,
	headers []string,
	authority string,
	stickySessionKeys []string,
	callOptions *CallOptions,
	logOptions *LogOptions,
) error {
	if logOptions.ShowRequest {
		printPrettyJson(data)
	}

	path := fmt.Sprintf("/v2/models/%s/infer", resourceName)

	timeStart := time.Now().Unix()
	for i := 0; canContinueInfer(callOptions, i, timeStart); i++ {
		res, err := ic.httpCall(resourceName, path, data, callOptions.InferType, logOptions.ShowHeaders, headers, authority, stickySessionKeys)
		if err != nil {
			if callOptions.Iterations == 1 && callOptions.Seconds == 0 {
				return err
			}
			continue
		}

		v2InferResponse := V2InferenceResponse{}
		err = json.Unmarshal(res, &v2InferResponse)
		if err != nil {
			return err
		}

		if callOptions.Iterations == 1 && callOptions.Seconds == 0 {
			if logOptions.ShowResponse {
				printPrettyJson(res)
			}
		}
	}

	if callOptions.Iterations > 1 || callOptions.Seconds > 0 {
		fmt.Printf("Success: %v\n", ic.counts)

	}
	if len(ic.errors) > 0 {
		fmt.Printf("Errors: %v\n", ic.errors)
	}
	return nil
}

func getDataSize(shape []int64) int64 {
	tot := int64(1)
	for _, dim := range shape {
		tot = tot * dim
	}
	return tot
}

func isNilOutputContents(contents *v2_dataplane.InferTensorContents) bool {
	if contents == nil {
		return true
	} else {
		if contents.BoolContents == nil &&
			contents.BytesContents == nil &&
			contents.IntContents == nil &&
			contents.Int64Contents == nil &&
			contents.Fp32Contents == nil &&
			contents.Fp64Contents == nil &&
			contents.UintContents == nil &&
			contents.Uint64Contents == nil {
			return true
		}
	}
	return false
}

func updateResponseFromRawContents(res *v2_dataplane.ModelInferResponse) error {
	outputIdx := 0
	for _, rawOutput := range res.RawOutputContents {
		contents := &v2_dataplane.InferTensorContents{}
		for ; outputIdx < len(res.Outputs); outputIdx++ {
			if isNilOutputContents(res.Outputs[outputIdx].Contents) {
				break
			}
		}
		if outputIdx == len(res.Outputs) {
			return fmt.Errorf("Ran out of output contents to fill raw contents of length %d", len(res.RawOutputContents))
		}
		output := res.Outputs[outputIdx]
		output.Contents = contents
		var err error
		switch output.Datatype {
		case tyBool:
			output.Contents.BoolContents = make([]bool, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.BoolContents)
		case tyUint8, tyUint16, tyUint32:
			output.Contents.UintContents = make([]uint32, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.UintContents)
		case tyUint64:
			output.Contents.Uint64Contents = make([]uint64, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.Uint64Contents)
		case tyInt8, tyInt16, tyInt32:
			output.Contents.IntContents = make([]int32, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.IntContents)
		case tyInt64:
			output.Contents.Int64Contents = make([]int64, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.Int64Contents)
		case tyFp16, tyFp32:
			output.Contents.Fp32Contents = make([]float32, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.Fp32Contents)
		case tyFp64:
			output.Contents.Fp64Contents = make([]float64, getDataSize(output.Shape))
			err = binary.Read(bytes.NewBuffer(rawOutput), binary.LittleEndian, &output.Contents.Fp64Contents)
		case tyBytes:
			output.Contents.BytesContents = convertRawBytesToByteContents(rawOutput)
		}
		if err != nil {
			return err
		}
	}
	// Clear the raw contents now we have copied
	res.RawOutputContents = nil
	return nil
}

// Follows Triton client
// see https://github.com/triton-inference-server/client/blob/6cc412c50ca4282cec6e9f62b3c2781be433dcc6/src/python/library/tritonclient/utils/__init__.py#L246-L273
func convertRawBytesToByteContents(raw []byte) [][]byte {
	var result [][]byte
	for offset := uint32(0); offset < uint32(len(raw)); {
		dataLen := binary.LittleEndian.Uint32(raw[offset : offset+4])
		offset += 4
		data := raw[offset : offset+dataLen]
		result = append(result, data)
		offset += dataLen
	}
	return result
}

func updateRequestFromRawContents(res *v2_dataplane.ModelInferRequest) error {
	if len(res.RawInputContents) == len(res.Inputs) {
		for idx, inputs := range res.Inputs {
			contents := &v2_dataplane.InferTensorContents{}
			inputs.Contents = contents
			var err error
			switch inputs.Datatype {
			case tyBool:
				inputs.Contents.BoolContents = make([]bool, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.BoolContents)
			case tyUint8, tyUint16, tyUint32:
				inputs.Contents.UintContents = make([]uint32, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.UintContents)
			case tyUint64:
				inputs.Contents.Uint64Contents = make([]uint64, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.Uint64Contents)
			case tyInt8, tyInt16, tyInt32:
				inputs.Contents.IntContents = make([]int32, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.IntContents)
			case tyInt64:
				inputs.Contents.Int64Contents = make([]int64, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.Int64Contents)
			case tyFp16, tyFp32:
				inputs.Contents.Fp32Contents = make([]float32, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.Fp32Contents)
			case tyFp64:
				inputs.Contents.Fp64Contents = make([]float64, getDataSize(inputs.Shape))
				err = binary.Read(bytes.NewBuffer(res.RawInputContents[idx]), binary.LittleEndian, &inputs.Contents.Fp64Contents)
			case tyBytes:
				inputs.Contents.BytesContents = make([][]byte, 1)
				inputs.Contents.BytesContents[0] = res.RawInputContents[idx]
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ic *InferenceClient) InferGrpc(
	resourceName string,
	data []byte,
	headers []string,
	authority string,
	stickySessionKeys []string,
	callOptions *CallOptions,
	logOptions *LogOptions,
) error {
	hs, err := validateHeaders(headers)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx = addHeadersToContext(ctx, hs)
	ctx = addStickySessionToContext(ctx, stickySessionKeys)
	ctx = addSeldonModelHeaderToContext(ctx, callOptions.InferType, resourceName)

	req := &v2_dataplane.ModelInferRequest{}
	err = protojson.Unmarshal(data, req)
	if err != nil {
		return err
	}
	req.ModelName = resourceName

	conn, err := ic.newGRPCConnection(authority, logOptions)
	if err != nil {
		return err
	}

	grpcClient := v2_dataplane.NewGRPCInferenceServiceClient(conn)

	timeStart := time.Now().Unix()
	for i := 0; canContinueInfer(callOptions, i, timeStart); i++ {
		var header metadata.MD
		res, err := grpcClient.ModelInfer(ctx, req, grpc.Header(&header))
		if err != nil {
			if callOptions.Iterations == 1 && callOptions.Seconds == 0 {
				return err
			}
			if e, ok := status.FromError(err); ok {
				ic.updateErrors(int(e.Code()), header.Get(SeldonRouteHeader))
			}
			continue
		}

		if callOptions.Iterations == 1 && callOptions.Seconds == 0 {
			if logOptions.ShowResponse {
				err := updateResponseFromRawContents(res)
				if err != nil {
					return err
				}
				printProto(res)
			}
		} else {
			ic.updateSummary(header.Get(SeldonRouteHeader))
		}

		_, err = saveStickySessionKeyGrpc(header)
		if err != nil {
			return err
		}
	}

	if callOptions.Iterations > 1 || callOptions.Seconds > 0 {
		fmt.Printf("Success: %v\n", ic.counts)
		if len(ic.errors) > 0 {
			fmt.Printf("Errors: %v\n", ic.errors)
		}
	}
	return nil
}

func validateHeaders(headers []string) (map[string]string, error) {
	hs := make(map[string]string, len(headers))

	for _, header := range headers {
		parts := strings.Split(header, HeaderSeparator)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Badly formed header %s: use key%sval", header, HeaderSeparator)
		}

		err := rejectVirtualHostHeader(parts[0])
		if err != nil {
			return nil, err
		}

		hs[parts[0]] = parts[1]
	}

	return hs, nil
}

func rejectVirtualHostHeader(header string) error {
	normalised := strings.ToLower(strings.TrimSpace(header))
	if "host" == normalised || "authority" == normalised || ":authority" == normalised {
		return fmt.Errorf("Setting %s via headers is not supported, please use '--authority' instead", header)
	}
	return nil
}

func addHeadersToContext(ctx context.Context, headers map[string]string) context.Context {
	for k, v := range headers {
		ctx = metadata.AppendToOutgoingContext(ctx, k, v)
	}
	return ctx
}

func addStickySessionToContext(ctx context.Context, stickySessionKeys []string) context.Context {
	for _, k := range stickySessionKeys {
		ctx = metadata.AppendToOutgoingContext(ctx, SeldonRouteHeader, k)
	}
	return ctx
}

func addSeldonModelHeaderToContext(ctx context.Context, inferType InferType, resourceName string) context.Context {
	var headerValue string
	switch inferType {
	case InferModel:
		headerValue = resourceName
	case InferPipeline:
		headerValue = resourceName + "." + SeldonPipelineHeader
	}

	return metadata.AppendToOutgoingContext(ctx, SeldonModelHeader, headerValue)
}

func (ic *InferenceClient) Infer(
	modelName string,
	data []byte,
	headers []string,
	authority string,
	callOptions *CallOptions,
	logOptions *LogOptions,
) error {
	var stickySessionKeys []string
	var err error
	if callOptions.StickySession {
		stickySessionKeys, err = getStickySessionKeys()
		if err != nil {
			return err
		}
	}
	switch callOptions.InferProtocol {
	case "rest":
		return ic.InferRest(modelName, data, headers, authority, stickySessionKeys, callOptions, logOptions)
	case "grpc":
		return ic.InferGrpc(modelName, data, headers, authority, stickySessionKeys, callOptions, logOptions)
	default:
		return fmt.Errorf("Unknown infer mode - needs to be grpc or rest")
	}
}
